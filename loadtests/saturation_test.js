import http from "k6/http";
import { check } from "k6";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const SHORT_CODE = __ENV.SHORT_CODE || "test";

// Step-Up Stress Test: Gradually increases RPS to find the system's Breaking / Saturation Point
export const options = {
  summaryTrendStats: ["avg", "min", "med", "p(90)", "p(95)", "p(99)", "max"],

  scenarios: {
    saturation_curve: {
      executor: "ramping-arrival-rate",
      startRate: 1000,
      timeUnit: "1s",
      preAllocatedVUs: 1000,
      maxVUs: 10000,
      stages: [
        { duration: "20s", target: 2000 },  // Stage 1: Warmup & Light Load (2,000 RPS)
        { duration: "30s", target: 10000 }, // Stage 2: Moderate Load (10,000 RPS)
        { duration: "30s", target: 25000 }, // Stage 3: High Scale (25,000 RPS)
        { duration: "30s", target: 50000 }, // Stage 4: Extreme Saturation (50,000 RPS)
        { duration: "20s", target: 0 },     // Stage 5: Cooldown / Recovery
      ],
    },
  },

  thresholds: {
    // SRE SLO: Less than 1% errors under normal load
    http_req_failed: ["rate<0.01"],
    // SRE SLO: 95% of redirects must resolve in under 5ms
    http_req_duration: ["p(95)<5", "p(99)<20"],
  },
};

export default function () {
  const res = http.get(`${BASE_URL}/${SHORT_CODE}`, {
    redirects: 0,
    tags: { endpoint: "redirect_saturation" },
  });

  check(res, {
    "status is 302": (r) => r.status === 302,
    "has Location header": (r) => Boolean(r.headers.Location),
  });
}
