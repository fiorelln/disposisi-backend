// public/js/api.js

const BASE_URL = "http://localhost:7000/auth";

async function apiPost(endpoint, data) {
    const res = await fetch(`${BASE_URL}/${endpoint}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data)
    });

    const body = await res.json().catch(() => ({}));

    return {
        success: res.status === 200,
        statusCode: res.status,
        data: body,
    };
}
