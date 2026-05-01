package services

import (
	"bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"test-iq-ku/apps/api/internal/config"
	"test-iq-ku/apps/api/internal/models"
)

const (
	proUpgradeProductCode   = "PRO_UPGRADE"
	maxUpgradeProductCode   = "MAX_UPGRADE"
	proUpgradeDefaultAmount = 40000
	maxUpgradeDefaultAmount = 50000
	proUpgradeSubmitLimit   = 10
	maxUpgradeSubmitLimit   = 0
)

type CreateUpgradePaymentResult struct {
	OrderID     string `json:"orderId"`
	SnapToken   string `json:"snapToken"`
	RedirectURL string `json:"redirectUrl"`
	Amount      int    `json:"amount"`
	Status      string `json:"status"`
	AccountType string `json:"accountType"`
}

type CreateUpgradePaymentInput struct {
	AccountType models.AccountType `json:"accountType"`
}

type ConfirmPaymentInput struct {
	OrderID string `json:"orderId"`
}

type PaymentStatusResult struct {
	OrderID           string               `json:"orderId"`
	AccountType       models.AccountType   `json:"accountType"`
	PaymentStatus     models.PaymentStatus `json:"paymentStatus"`
	TransactionStatus string               `json:"transactionStatus"`
}

type UpgradePlanView struct {
	AccountType       models.AccountType `json:"accountType"`
	ProductCode       string             `json:"productCode"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	Amount            int                `json:"amount"`
	FormattedPrice    string             `json:"formattedPrice"`
	SubmitLimitPerDay int                `json:"submitLimitPerDay"`
	IsActive          bool               `json:"isActive"`
}

type UpgradePlanPricePayload struct {
	AccountType models.AccountType `json:"accountType"`
	Amount      int                `json:"amount"`
}

type UpdateUpgradePlanPricesInput struct {
	Plans []UpgradePlanPricePayload `json:"plans"`
}

type midtransSnapResponse struct {
	Token       string `json:"token"`
	RedirectURL string `json:"redirect_url"`
}

type midtransStatusResponse struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	TransactionStatus string `json:"transaction_status"`
	FraudStatus       string `json:"fraud_status"`
	TransactionID     string `json:"transaction_id"`
	PaymentType       string `json:"payment_type"`
	SignatureKey      string `json:"signature_key"`
	StatusMessage     string `json:"status_message"`
}

type upgradeProduct struct {
	ProductCode string
	AccountType models.AccountType
	Amount      int
	OrderPrefix string
	ItemName    string
}

type PaymentService struct {
	db     *gorm.DB
	cfg    *config.Config
	client *http.Client
}

func NewPaymentService(db *gorm.DB, cfg *config.Config) *PaymentService {
	return &PaymentService{
		db:  db,
		cfg: cfg,
		client: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (s *PaymentService) ListUpgradePlans() ([]UpgradePlanView, error) {
	plans, err := s.listUpgradePlanModels(s.db)
	if err != nil {
		return nil, err
	}

	return serializeUpgradePlans(plans), nil
}

func (s *PaymentService) UpdateUpgradePlanPrices(input UpdateUpgradePlanPricesInput) ([]UpgradePlanView, error) {
	if len(input.Plans) == 0 {
		return nil, errors.New("minimal satu harga paket wajib dikirim")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	if err := s.ensureDefaultUpgradePlans(tx); err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, item := range input.Plans {
		accountType := canonicalAccountType(item.AccountType)
		if accountType != models.AccountTypePro && accountType != models.AccountTypeMax {
			tx.Rollback()
			return nil, errors.New("paket membership tidak valid")
		}
		if item.Amount <= 0 {
			tx.Rollback()
			return nil, fmt.Errorf("harga %s harus lebih dari 0", accountType)
		}

		if err := tx.Model(&models.MembershipPlan{}).
			Where("account_type = ?", accountType).
			Update("amount", item.Amount).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	var plans []models.MembershipPlan
	if err := tx.Where("account_type IN ?", []models.AccountType{models.AccountTypePro, models.AccountTypeMax}).
		Find(&plans).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return serializeUpgradePlans(orderUpgradePlans(plans)), nil
}

func (s *PaymentService) CreateUpgradePayment(user models.User, input CreateUpgradePaymentInput) (*CreateUpgradePaymentResult, error) {
	product, err := s.resolveUpgradeProduct(input.AccountType)
	if err != nil {
		return nil, err
	}
	currentAccountType := canonicalAccountType(user.AccountType)
	if currentAccountType == product.AccountType {
		return nil, fmt.Errorf("akun Anda sudah %s", product.AccountType)
	}
	if currentAccountType == models.AccountTypeMax {
		return nil, errors.New("akun Anda sudah MAX")
	}
	if strings.TrimSpace(s.cfg.MidtransServerKey) == "" {
		return nil, errors.New("konfigurasi Midtrans belum lengkap")
	}

	var existing models.PaymentTransaction
	if err := s.db.
		Where("user_id = ? AND product_code = ? AND status IN ?", user.ID, product.ProductCode, []models.PaymentStatus{
			models.PaymentStatusInitiated,
			models.PaymentStatusPending,
		}).
		Order("id DESC").
		First(&existing).Error; err == nil {
		if existing.Amount != product.Amount {
			if err := s.db.Model(&existing).Update("status", models.PaymentStatusCanceled).Error; err != nil {
				return nil, err
			}
		} else {
			return &CreateUpgradePaymentResult{
				OrderID:     existing.OrderID,
				SnapToken:   existing.SnapToken,
				RedirectURL: existing.RedirectURL,
				Amount:      existing.Amount,
				Status:      string(existing.Status),
				AccountType: string(product.AccountType),
			}, nil
		}
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	orderID := fmt.Sprintf("%s-%d-%d", product.OrderPrefix, user.ID, time.Now().Unix())
	payload := fiberMap{
		"transaction_details": fiberMap{
			"order_id":     orderID,
			"gross_amount": product.Amount,
		},
		"credit_card": fiberMap{
			"secure": true,
		},
		"item_details": []fiberMap{
			{
				"id":       product.ProductCode,
				"price":    product.Amount,
				"quantity": 1,
				"name":     product.ItemName,
			},
		},
		"customer_details": fiberMap{
			"first_name": user.Name,
			"email":      user.Email,
		},
	}

	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	responseBody, err := s.doMidtransRequest(http.MethodPost, s.snapTransactionURL(), rawPayload)
	if err != nil {
		return nil, err
	}

	var snapRes midtransSnapResponse
	if err := json.Unmarshal(responseBody, &snapRes); err != nil {
		return nil, err
	}
	if strings.TrimSpace(snapRes.Token) == "" {
		return nil, errors.New("snap token Midtrans tidak tersedia")
	}

	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	record := models.PaymentTransaction{
		UserID:            user.ID,
		OrderID:           orderID,
		ProductCode:       product.ProductCode,
		Amount:            product.Amount,
		Status:            models.PaymentStatusInitiated,
		SnapToken:         snapRes.Token,
		RedirectURL:       snapRes.RedirectURL,
		TransactionStatus: "created",
		Metadata:          datatypes.JSON(rawPayload),
	}

	if err := tx.Create(&record).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &CreateUpgradePaymentResult{
		OrderID:     orderID,
		SnapToken:   snapRes.Token,
		RedirectURL: snapRes.RedirectURL,
		Amount:      product.Amount,
		Status:      string(record.Status),
		AccountType: string(product.AccountType),
	}, nil
}

func (s *PaymentService) ConfirmPayment(user models.User, input ConfirmPaymentInput) (*PaymentStatusResult, error) {
	if strings.TrimSpace(input.OrderID) == "" {
		return nil, errors.New("order id wajib diisi")
	}

	var payment models.PaymentTransaction
	if err := s.db.Where("user_id = ? AND order_id = ?", user.ID, strings.TrimSpace(input.OrderID)).First(&payment).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaksi pembayaran tidak ditemukan")
		}
		return nil, err
	}

	status, err := s.fetchTransactionStatus(payment.OrderID)
	if err != nil {
		return nil, err
	}

	result, err := s.applyMidtransStatus(status)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *PaymentService) HandleMidtransNotification(rawBody []byte) (*PaymentStatusResult, error) {
	var payload midtransStatusResponse
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, errors.New("payload notifikasi Midtrans tidak valid")
	}
	if !s.isValidMidtransSignature(payload) {
		return nil, errors.New("signature Midtrans tidak valid")
	}
	return s.applyMidtransStatus(payload)
}

func (s *PaymentService) applyMidtransStatus(payload midtransStatusResponse) (*PaymentStatusResult, error) {
	tx := s.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var payment models.PaymentTransaction
	if err := tx.Where("order_id = ?", strings.TrimSpace(payload.OrderID)).First(&payment).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaksi pembayaran tidak ditemukan")
		}
		return nil, err
	}

	var user models.User
	if err := tx.Where("id = ?", payment.UserID).First(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	nextStatus := mapMidtransTransactionStatus(payload.TransactionStatus, payload.FraudStatus)
	now := time.Now()
	payment.Status = nextStatus
	payment.TransactionStatus = strings.TrimSpace(payload.TransactionStatus)
	payment.FraudStatus = strings.TrimSpace(payload.FraudStatus)
	payment.PaymentType = strings.TrimSpace(payload.PaymentType)
	payment.MidtransTransactionID = strings.TrimSpace(payload.TransactionID)

	rawStatus, err := json.Marshal(payload)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	payment.Metadata = datatypes.JSON(rawStatus)

	if nextStatus == models.PaymentStatusPaid && payment.PaidAt == nil {
		payment.PaidAt = &now
	}

	if err := tx.Save(&payment).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if nextStatus == models.PaymentStatusPaid {
		targetAccountType := s.accountTypeFromProductCode(tx, payment.ProductCode)
		if canonicalAccountType(user.AccountType) != targetAccountType {
			user.AccountType = targetAccountType
		}
	}

	if nextStatus == models.PaymentStatusPaid {
		if err := tx.Model(&user).Update("account_type", user.AccountType).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &PaymentStatusResult{
		OrderID:           payment.OrderID,
		AccountType:       canonicalAccountType(user.AccountType),
		PaymentStatus:     payment.Status,
		TransactionStatus: payment.TransactionStatus,
	}, nil
}

func (s *PaymentService) resolveUpgradeProduct(accountType models.AccountType) (upgradeProduct, error) {
	switch canonicalAccountType(accountType) {
	case models.AccountTypePro:
		return s.resolveUpgradeProductFromPlan(models.AccountTypePro)
	case models.AccountTypeMax:
		return s.resolveUpgradeProductFromPlan(models.AccountTypeMax)
	default:
		return upgradeProduct{}, errors.New("paket upgrade tidak valid")
	}
}

func (s *PaymentService) resolveUpgradeProductFromPlan(accountType models.AccountType) (upgradeProduct, error) {
	if err := s.ensureDefaultUpgradePlans(s.db); err != nil {
		return upgradeProduct{}, err
	}

	var plan models.MembershipPlan
	if err := s.db.Where("account_type = ? AND is_active = ?", accountType, true).First(&plan).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return upgradeProduct{}, errors.New("paket upgrade tidak aktif")
		}
		return upgradeProduct{}, err
	}
	if plan.Amount <= 0 {
		return upgradeProduct{}, fmt.Errorf("harga %s belum valid", accountType)
	}

	return upgradeProduct{
		ProductCode: plan.ProductCode,
		AccountType: canonicalAccountType(plan.AccountType),
		Amount:      plan.Amount,
		OrderPrefix: strings.TrimSuffix(plan.ProductCode, "_UPGRADE"),
		ItemName:    plan.Name,
	}, nil
}

func (s *PaymentService) accountTypeFromProductCode(tx *gorm.DB, productCode string) models.AccountType {
	var plan models.MembershipPlan
	if err := tx.Where("product_code = ?", strings.TrimSpace(productCode)).First(&plan).Error; err == nil {
		return canonicalAccountType(plan.AccountType)
	}
	return fallbackAccountTypeFromProductCode(productCode)
}

func fallbackAccountTypeFromProductCode(productCode string) models.AccountType {
	switch strings.TrimSpace(productCode) {
	case proUpgradeProductCode:
		return models.AccountTypePro
	case maxUpgradeProductCode:
		return models.AccountTypeMax
	default:
		return models.AccountTypeFree
	}
}

func (s *PaymentService) listUpgradePlanModels(db *gorm.DB) ([]models.MembershipPlan, error) {
	if err := s.ensureDefaultUpgradePlans(db); err != nil {
		return nil, err
	}

	var plans []models.MembershipPlan
	if err := db.Where("account_type IN ?", []models.AccountType{models.AccountTypePro, models.AccountTypeMax}).
		Find(&plans).Error; err != nil {
		return nil, err
	}

	return orderUpgradePlans(plans), nil
}

func (s *PaymentService) ensureDefaultUpgradePlans(db *gorm.DB) error {
	defaults := defaultUpgradePlanModels()
	for _, item := range defaults {
		var count int64
		if err := db.Model(&models.MembershipPlan{}).
			Where("account_type = ?", item.AccountType).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
}

func defaultUpgradePlanModels() []models.MembershipPlan {
	return []models.MembershipPlan{
		{
			AccountType:       models.AccountTypePro,
			ProductCode:       proUpgradeProductCode,
			Name:              "Paket Pro",
			Description:       "Semua fitur terbuka dengan batas 10 submit IQ dan 10 submit SKB per hari.",
			Amount:            proUpgradeDefaultAmount,
			SubmitLimitPerDay: proUpgradeSubmitLimit,
			IsActive:          true,
		},
		{
			AccountType:       models.AccountTypeMax,
			ProductCode:       maxUpgradeProductCode,
			Name:              "Paket Max",
			Description:       "Full akses tanpa batas submit harian untuk IQ dan SKB.",
			Amount:            maxUpgradeDefaultAmount,
			SubmitLimitPerDay: maxUpgradeSubmitLimit,
			IsActive:          true,
		},
	}
}

func orderUpgradePlans(plans []models.MembershipPlan) []models.MembershipPlan {
	byType := make(map[models.AccountType]models.MembershipPlan, len(plans))
	for _, plan := range plans {
		byType[canonicalAccountType(plan.AccountType)] = plan
	}

	ordered := make([]models.MembershipPlan, 0, len(plans))
	for _, accountType := range []models.AccountType{models.AccountTypePro, models.AccountTypeMax} {
		if plan, ok := byType[accountType]; ok {
			ordered = append(ordered, plan)
		}
	}

	return ordered
}

func serializeUpgradePlans(plans []models.MembershipPlan) []UpgradePlanView {
	response := make([]UpgradePlanView, 0, len(plans))
	for _, plan := range plans {
		response = append(response, UpgradePlanView{
			AccountType:       canonicalAccountType(plan.AccountType),
			ProductCode:       plan.ProductCode,
			Name:              plan.Name,
			Description:       plan.Description,
			Amount:            plan.Amount,
			FormattedPrice:    formatRupiah(plan.Amount),
			SubmitLimitPerDay: plan.SubmitLimitPerDay,
			IsActive:          plan.IsActive,
		})
	}
	return response
}

func formatRupiah(amount int) string {
	if amount < 0 {
		return "-Rp" + formatThousands(-amount)
	}
	return "Rp" + formatThousands(amount)
}

func formatThousands(value int) string {
	raw := fmt.Sprintf("%d", value)
	if len(raw) <= 3 {
		return raw
	}

	remainder := len(raw) % 3
	if remainder == 0 {
		remainder = 3
	}

	var builder strings.Builder
	builder.WriteString(raw[:remainder])
	for index := remainder; index < len(raw); index += 3 {
		builder.WriteString(".")
		builder.WriteString(raw[index : index+3])
	}
	return builder.String()
}

func (s *PaymentService) fetchTransactionStatus(orderID string) (midtransStatusResponse, error) {
	responseBody, err := s.doMidtransRequest(http.MethodGet, s.statusURL(orderID), nil)
	if err != nil {
		return midtransStatusResponse{}, err
	}

	var status midtransStatusResponse
	if err := json.Unmarshal(responseBody, &status); err != nil {
		return midtransStatusResponse{}, err
	}

	return status, nil
}

func (s *PaymentService) doMidtransRequest(method string, endpoint string, body []byte) ([]byte, error) {
	var requestBody io.Reader
	if len(body) > 0 {
		requestBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, endpoint, requestBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	encodedKey := base64.StdEncoding.EncodeToString([]byte(s.cfg.MidtransServerKey + ":"))
	req.Header.Set("Authorization", "Basic "+encodedKey)

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("midtrans request gagal: %s", strings.TrimSpace(string(responseBody)))
	}

	return responseBody, nil
}

func (s *PaymentService) snapTransactionURL() string {
	if s.cfg.MidtransProduction {
		return "https://app.midtrans.com/snap/v1/transactions"
	}
	return "https://app.sandbox.midtrans.com/snap/v1/transactions"
}

func (s *PaymentService) statusURL(orderID string) string {
	baseURL := "https://api.sandbox.midtrans.com/v2"
	if s.cfg.MidtransProduction {
		baseURL = "https://api.midtrans.com/v2"
	}
	return fmt.Sprintf("%s/%s/status", baseURL, orderID)
}

func (s *PaymentService) isValidMidtransSignature(payload midtransStatusResponse) bool {
	if strings.TrimSpace(s.cfg.MidtransServerKey) == "" {
		return false
	}
	expected := buildMidtransSignature(payload.OrderID, payload.StatusCode, payload.GrossAmount, s.cfg.MidtransServerKey)
	return strings.EqualFold(expected, strings.TrimSpace(payload.SignatureKey))
}

func buildMidtransSignature(orderID string, statusCode string, grossAmount string, serverKey string) string {
	sum := sha512.Sum512([]byte(strings.TrimSpace(orderID) + strings.TrimSpace(statusCode) + strings.TrimSpace(grossAmount) + strings.TrimSpace(serverKey)))
	return hex.EncodeToString(sum[:])
}

func mapMidtransTransactionStatus(transactionStatus string, fraudStatus string) models.PaymentStatus {
	switch strings.ToLower(strings.TrimSpace(transactionStatus)) {
	case "settlement":
		return models.PaymentStatusPaid
	case "capture":
		if strings.EqualFold(strings.TrimSpace(fraudStatus), "accept") || strings.TrimSpace(fraudStatus) == "" {
			return models.PaymentStatusPaid
		}
		return models.PaymentStatusPending
	case "pending":
		return models.PaymentStatusPending
	case "expire":
		return models.PaymentStatusExpired
	case "cancel":
		return models.PaymentStatusCanceled
	case "deny", "failure":
		return models.PaymentStatusFailed
	default:
		return models.PaymentStatusPending
	}
}

type fiberMap map[string]any
