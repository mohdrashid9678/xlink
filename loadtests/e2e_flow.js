import http from "k6/http";
import { check, sleep, group } from "k6";
import { randomString } from "https://jslib.k6.io/k6-utils/1.4.0/index.js";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const VUS = parseInt(__ENV.VUS || "20", 10);
const DURATION = __ENV.DURATION || "30s";

export const options = {
  summaryTrendStats: ["avg", "min", "med", "p(90)", "p(95)", "p(99)", "max"],
  stages: [
    { duration: "5s", target: VUS },     // Ramp up
    { duration: DURATION, target: VUS },  // Steady state
    { duration: "5s", target: 0 },       // Ramp down
  ],
  thresholds: {
    // True server failures (HTTP 5xx) must be < 1%
    "http_req_failed{expected:true}": ["rate<0.01"],
    // SRE SLO: Redirect reads should be fast (p95 < 50ms under heavy DB write mix)
    "http_req_duration{type:redirect}": ["p(95)<50", "p(99)<100"],
    // SRE SLO: Auth & URL creation operations (including bcrypt hashing)
    "http_req_duration{type:api}": ["p(95)<600", "p(99)<900"],
  },
};

export function setup() {
  const healthRes = http.get(`${BASE_URL}/healthz`);
  check(healthRes, {
    "system is healthy": (r) => r.status === 200,
  });
  return { startTime: new Date().toISOString() };
}

export default function () {
  const uniqueId = `${__VU}_${__ITER}_${Date.now()}_${randomString(4)}`;
  const email = `testuser_${uniqueId}@example.com`;
  const password = "Password123!";
  const name = `LoadTester ${uniqueId}`;

  let token = "";
  const createdShortCodes = [];

  // ==========================================
  // STEP 1: User Registration
  // ==========================================
  group("1. User Signup", function () {
    const payload = JSON.stringify({
      email: email,
      password: password,
      name: name,
    });

    const res = http.post(`${BASE_URL}/api/v1/auth/register`, payload, {
      headers: { "Content-Type": "application/json" },
      tags: { type: "api", endpoint: "auth_register", expected: "true" },
    });

    const passed = check(res, {
      "signup status is 201": (r) => r.status === 201,
      "has access token": (r) => Boolean(r.json("data.tokens.access_token")),
    });

    if (passed) {
      token = res.json("data.tokens.access_token");
    }
  });

  if (!token) return;

  const authHeaders = {
    "Content-Type": "application/json",
    Authorization: `Bearer ${token}`,
  };

  // ==========================================
  // STEP 2: Batch URL Creation
  // ==========================================
  group("2. Batch URL Creation", function () {
    const urlsToCreate = 10;

    for (let i = 1; i <= urlsToCreate; i++) {
      const longURL = `https://example.com/products/item-${uniqueId}-${i}?ref=k6_loadtest`;
      const createPayload = JSON.stringify({
        long_url: longURL,
      });

      const res = http.post(`${BASE_URL}/api/v1/urls`, createPayload, {
        headers: authHeaders,
        tags: { type: "api", endpoint: "url_create", expected: "true" },
      });

      const success = check(res, {
        "url create is 201": (r) => r.status === 201,
        "has short_code": (r) => Boolean(r.json("data.short_code")),
      });

      if (success) {
        createdShortCodes.push(res.json("data.short_code"));
      }
    }
  });

  // ==========================================
  // STEP 3: High-Speed Redirect Reads
  // ==========================================
  group("3. Redirect Read Operations", function () {
    if (createdShortCodes.length === 0) return;

    for (let i = 0; i < 20; i++) {
      const shortCode = createdShortCodes[i % createdShortCodes.length];

      const res = http.get(`${BASE_URL}/${shortCode}`, {
        redirects: 0,
        tags: { type: "redirect", endpoint: "url_redirect", expected: "true" },
      });

      check(res, {
        "redirect is 302 Found": (r) => r.status === 302,
        "location header present": (r) => Boolean(r.headers.Location),
      });
    }

    // Intentional 404 test: Tag with expected: false so it doesn't fail server error thresholds
    const missingRes = http.get(`${BASE_URL}/non_existent_${randomString(6)}`, {
      redirects: 0,
      tags: { type: "redirect", endpoint: "url_404", expected: "false" },
    });

    check(missingRes, {
      "missing URL returns 404": (r) => r.status === 404,
    });
  });

  sleep(0.3);
}
