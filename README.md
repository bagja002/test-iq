# Test IQ Ku

Monorepo untuk platform test IQ berbasis `Next.js` dan `Go Fiber v3`.

## Struktur

- `apps/web`: aplikasi frontend Next.js + shadcn/ui
- `apps/api`: REST API Go Fiber v3 + MySQL + GORM
- `packages/openapi`: kontrak OpenAPI dan tipe shared untuk frontend
- `infra`: file deploy awal seperti Nginx

## Menjalankan Frontend

```bash
npm install
npm run dev:web
```

## Menjalankan Backend

```bash
cd apps/api
go run ./cmd/server
```

## Environment

Lihat file `.env.example` di `apps/web` dan `apps/api`.
