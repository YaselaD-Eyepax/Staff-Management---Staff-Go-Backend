import http from "k6/http";
import { check } from "k6";

export let options = {
  vus: 1, // start low
  duration: "10s",
};

const BASE_URL = "http://localhost:8085/api/v1";

export default function () {
  const payload = JSON.stringify({
    title: `Scheduled System Maintenance ${Math.random()}`,
    summary: "Planned downtime for the Staff Management System.",
    body: "A system-wide maintenance will take place on Saturday from 10 PM to 1 AM. During this window, logins, room bookings, shift updates, and leave submissions will be temporarily unavailable.",
    attachments: [],
    tags: ["maintenance", "system", "downtime"],
    created_by: "7fcd6769-617b-4454-b6c7-2d15cddf9c11",
    scheduled_at: "2025-12-01T18:30:00Z",
  });

  const params = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const res = http.post(`${BASE_URL}/events`, payload, params);

  check(res, {
    "event created (201)": (r) => r.status === 201,
  });
}
