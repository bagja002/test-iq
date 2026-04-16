import http from "k6/http"
import { check, sleep } from "k6"

export const options = {
  scenarios: {
    login_start_submit: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "30s", target: 25 },
        { duration: "1m", target: 100 },
        { duration: "30s", target: 0 },
      ],
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<700"],
    http_req_failed: ["rate<0.02"],
  },
}

const baseUrl = __ENV.BASE_URL || "http://localhost"

export default function () {
  const jar = http.cookieJar()
  const email = __ENV.EMAIL || "user@testiq.local"
  const password = __ENV.PASSWORD || "user123"

  const loginResponse = http.post(
    `${baseUrl}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    {
      headers: { "Content-Type": "application/json" },
      jar,
    },
  )

  check(loginResponse, {
    "login ok": (res) => res.status === 200,
  })

  const startResponse = http.post(`${baseUrl}/api/v1/test-attempts/start`, null, { jar })
  check(startResponse, {
    "start ok": (res) => res.status === 200,
  })

  const attemptId = startResponse.json("attemptId")
  if (!attemptId) {
    sleep(1)
    return
  }

  const detailResponse = http.get(`${baseUrl}/api/v1/test-attempts/${attemptId}`, { jar })
  const questions = detailResponse.json("questions") || []

  if (questions.length) {
    const firstQuestion = questions[0]
    http.put(
      `${baseUrl}/api/v1/test-attempts/${attemptId}/answers`,
      JSON.stringify({
        answers: [
          {
            attemptQuestionId: firstQuestion.id,
            selectedOptionKey: firstQuestion.options[0].key,
          },
        ],
      }),
      {
        headers: { "Content-Type": "application/json" },
        jar,
      },
    )
  }

  http.post(`${baseUrl}/api/v1/test-attempts/${attemptId}/submit`, null, { jar })
  sleep(1)
}
