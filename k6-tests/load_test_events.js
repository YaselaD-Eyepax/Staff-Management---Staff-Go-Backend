import http from "k6/http";
import { sleep, check } from "k6";

export let options = {
  vus: 50, // 50 users
  duration: "30s", // for 30 seconds
};

const BASE_URL = "http://localhost:8085/api/v1";

export default function () {
  const res = http.get(`${BASE_URL}/events`);

  check(res, {
    "status is 200": (r) => r.status === 200,
    "response time < 500ms": (r) => r.timings.duration < 500,
  });

  sleep(1);
}
