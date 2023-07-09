const SERVER_URL = "http://gateway:3333";
const http = require('k6/http');

const MAPPER_URL = SERVER_URL + "/shorten";
const { check, sleep } = require('k6');

export const options = {
    // Key configurations for avg load test in this section
    stages: [
        { duration: '5m', target: 10 }, // traffic ramp-up from 1 to 100 users over 5 minutes.
        { duration: '30m', target: 10 }, // stay at 100 users for 10 minutes
        { duration: '5m', target: 0 }, // ramp-down to 0 users
    ],
};

export function setup(){
    // Sleep 20 seconds to allow the server to start
    sleep(20);
}

export default function () {
    const originalUrl = "https://google.com";

    const body = {
        url: originalUrl
    }

    const res = http.post(MAPPER_URL, JSON.stringify(body), {
        headers: {
            'Content-Type': 'application/json',
        },
    });
    check(res, {
        "status is 200": (r) => r.status === 200
    });
    sleep(0.5);
}