console.log("api.js loaded")
export async function postRequest(serverIP, data) {
    console.log("Sending:", JSON.stringify(data, null, 2))

    const res = await fetch(`http://${serverIP}`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify(data)
    });

    if (!res.ok) {
        throw new Error(`HTTP ${res.status}`);
    }

    return await res.json();
}
