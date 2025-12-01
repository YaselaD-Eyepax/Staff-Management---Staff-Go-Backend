import http from "k6/http";
import { sleep } from "k6";

export let options = {
  stages: [
    { duration: "10s", target: 20 }, // ramp up
    { duration: "20s", target: 50 }, // sustain
    { duration: "10s", target: 0 }, // ramp down
  ],
};

const BASE_URL = "http://localhost:8085/api/v1";

export default function () {
  http.get(`${BASE_URL}/events`);
  sleep(0.5);
}
