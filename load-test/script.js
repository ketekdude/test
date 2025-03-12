import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
    vus: 10, // Virtual Users
    duration: '30s', // Test duration
};

export default function () {
    let res = http.get('http://localhost:8080/test'); // Replace with your API URL

    check(res, {
        'status is 200': (r) => r.status === 200,
        'response time < 200ms': (r) => r.timings.duration < 200,
    });

    sleep(1); // Simulate real-world behavior
}