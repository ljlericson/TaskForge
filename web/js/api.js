export async function createJob(serverIP, jobRequest) {

    const response = await fetch(`http://${serverIP}/jobs/submit`, {

        method: "POST",

        headers: {
            "Content-Type": "application/json"
        },

        body: JSON.stringify(jobRequest)

    })

    if (!response.ok) {

        throw new Error("request failed")

    }

    return response.json()

}