import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  vus: 5, // moderate concurrency
  duration: "20s", // test window
  thresholds: {
    http_req_failed: ["rate<0.05"], // less than 5% errors allowed
    http_req_duration: ["p(95)<500"], // 95% under 500ms is healthy
  },
};

// 👇 You pass event ID as env variable
const EVENT_ID = "b65e6c81-8dac-439b-9254-4681ae986478";

const BASE_URL = "http://localhost:8085/api/v1";

export default function () {
  const payload = JSON.stringify({
    channels: ["fcm"], // safe for load tests; worker picks it up
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(
    `${BASE_URL}/events/${EVENT_ID}/broadcast`,
    payload,
    params
  );

  check(res, {
    "status is 200 OK": (r) => r.status === 200,
  });

  sleep(1); // Important to keep pressure reasonable
}
