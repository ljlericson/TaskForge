import { postRequest } from "./api.js"
import { showStatus } from "./ui.js"

document.addEventListener("DOMContentLoaded", () => {

    const form = document.getElementById("newJob")

    form.addEventListener("submit", async (e) => {
        console.log("asdas")
        e.preventDefault()

        const jobName = document.getElementById("jobName").value
        const timeoutSeconds = Number(document.getElementById("timeoutSeconds").value)
        const priority = Number(document.getElementById("priority").value)

        const jarURL = document.getElementById("jarURL").value
        const jarMainClass = document.getElementById("jarMainClass").value
        const jarArguments = document.getElementById("jarArguments").value

        const executionCores = Number(document.getElementById("executionCores").value)
        const executionMemory = Number(document.getElementById("executionMemory").value)

        const dataInput = document.getElementById("dataInput").value
        const dataOutput = document.getElementById("dataOutput").value

        const javaOpts = document.getElementById("javaOpts").value

        const serverIP = document.getElementById("serverIP").value
        const route = "/jobs/submit";
        const fullURL = `${serverIP}${route}`;


        // build JSON object
        const jobRequest = {

            jobName: jobName,

            jar: {
                url: jarURL,
                mainClass: jarMainClass
            },

            resources: {
                executionCores: executionCores,
                executionMemory: executionMemory
            },

            data: {
                input: [dataInput],
                output: dataOutput
            },

            arguments: jarArguments.split(" "),

            environment: {
                JAVA_OPTS: javaOpts
            },

            timeoutSeconds: timeoutSeconds,
            priority: priority
        }

        const button = document.getElementById("submitJob");

        try {
            const result = await postRequest(fullURL, jobRequest)
            console.log("job created:", result)
            showStatus(document.getElementById("submitJob"), "* job submitted", "success")

        } catch (err) {

            console.error("error:", err)
            showStatus(button, "* job failed to submit, error: " + err.message, "error")
        }

    })

})