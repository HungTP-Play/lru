const SERVER_URL = "http://gateway:3333";

const REDIRECT_URL = SERVER_URL + "/redirect";
const { check, sleep } = require('k6');
const http = require('k6/http');

export const options = {
    // Key configurations for avg load test in this section
    stages: [
        { duration: '5m', target: 10 }, // traffic ramp-up from 1 to 100 users over 5 minutes.
        { duration: '30m', target: 10 }, // stay at 100 users for 10 minutes
        { duration: '5m', target: 0 }, // ramp-down to 0 users
    ],
};

const base62 = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ";

const generateRandomString = (length) => {
    let result = '';
    for (let i = 0; i < length; i++) {
        result += base62.charAt(Math.floor(Math.random() * base62.length));
    }
    return result;
}

export function setup(){
    // Sleep 20 seconds to allow the server to start
    sleep(20);
}

export default function () {
    const shortenUrlId = generateRandomString(1);
    const shortUrl = 'http://localhost/' + shortenUrlId;

    const body = {
        url: shortUrl
    }

    const res = http.request('GET', REDIRECT_URL, JSON.stringify(body), {
        headers: {
            'Content-Type': 'application/json',
        },
    });
    check(res, {
        "status is 200": (r) => r.status === 200
    });
    sleep(0.5);
}