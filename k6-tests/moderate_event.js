import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  vus: 5, // reasonable write load
  duration: "20s", // moderate test duration

  thresholds: {
    http_req_failed: ["rate<0.05"], // <5% errors allowed
    http_req_duration: ["p(95)<800"], // 95% under 800ms
  },
};

// 🔥 Use event ID from --env EVENT_ID
const EVENT_ID = "ae8d20c5-5519-47bf-934e-773455370d71";

const BASE_URL = "http://localhost:8085/api/v1";

export default function () {
  const payload = JSON.stringify({
    status: "approve",
    moderator_id: "7fcd6769-617b-4454-b6c7-2d15cddf9c11", // your real admin ID
    notes: `Approved in k6 run at ..`,
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(
    `${BASE_URL}/events/${EVENT_ID}/moderate`,
    payload,
    params
  );

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(1); // keep writes safe
}
