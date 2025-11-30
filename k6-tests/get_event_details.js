import http from "k6/http";
import { check, sleep } from "k6";

export let options = {
  vus: 20,
  duration: "20s",
};

const BASE_URL = "http://localhost:8085/api/v1";
const TEST_EVENT_ID = "c87f85ad-eda7-473b-bd3d-86dbf7b8c30e";

export default function () {
  const res = http.get(`${BASE_URL}/events/${TEST_EVENT_ID}`);

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  sleep(1);
}
