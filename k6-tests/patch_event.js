import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  vus: 5,
  duration: "20s",

  thresholds: {
    http_req_failed: ["rate<0.05"], // <5% errors allowed
    http_req_duration: ["p(95)<800"], // 95% of requests < 800ms
  },
};

// 🔥 PUT YOUR REAL EVENT ID HERE
const EVENT_ID = "7d449fa3-e511-4e2b-ae0e-329addfb1e6f";

const BASE_URL = "http://localhost:8085/api/v1";

export default function () {
  const payload = JSON.stringify({
    title: `Updated Title ${Math.random()}`,
    summary: "Updated summary from k6",
    body: "Updated event body",
    attachments: [],
    tags: ["maintenance"],
    scheduled_at: "2025-12-05T18:30:00Z",
  });

  const params = {
    headers: { "Content-Type": "application/json" },
  };

  const res = http.patch(`${BASE_URL}/events/${EVENT_ID}`, payload, params);

  check(res, {
    "status is 200 OK": (r) => r.status === 200,
  });

  sleep(1); // keep load reasonable
}
