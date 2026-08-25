import http from "k6/http";
import { check } from "k6";

const BASE_URL = __ENV.BASE_URL || "http://localhost:8080";
const SHORT_CODE = __ENV.SHORT_CODE || "test";
const RATE = __ENV.RATE || 1000;

export const options = {
  summaryTrendStats: ["avg", "min", "med", "p(90)", "p(95)", "p(99)", "max"],

  scenarios: {
    redirects: {
      executor: "constant-arrival-rate",
      rate: RATE,
      timeUnit: "1s",
      duration: "10s",
      preAllocatedVUs: 1000,
      maxVUs: 5000,
    },
  },

  thresholds: {
    http_req_failed: ["rate<0.01"],
    http_req_duration: ["p(95)<100", "p(99)<200"],
  },
};

export default function () {
  const response = http.get(`${BASE_URL}/${SHORT_CODE}`, {
    redirects: 0,
    tags: {
      endpoint: "redirect",
    },
  });

  check(response, {
    "status is 302": (r) => r.status === 302,
    "has location header": (r) => Boolean(r.headers.Location),
  });
}
