document.addEventListener("DOMContentLoaded", () => {

    const form = document.getElementById("newJob")

    form.addEventListener("submit", async (e) => {

        // stop page reload
        e.preventDefault()

        // get values from inputs
        const jobName = document.getElementById("jobName").value
        const timeoutSeconds = Number(document.getElementById("timeoutSeconds").value)
        const priority = Number(document.getElementById("priority").value)

        const jarURL = document.getElementById("jarURL").value
        const jarMainClass = document.getElementById("jarMainClass").value
        const jarArguments = document.getElementById("jarArguments").value

        const executors = Number(document.getElementById("executors").value)
        const coresPerExecutor = Number(document.getElementById("coresPerExecutor").value)
        const memoryPerExecutorMB = Number(document.getElementById("memoryPerExecutorMB").value)

        const dataInput = document.getElementById("dataInput").value
        const dataOutput = document.getElementById("dataOutput").value

        const logLevel = document.getElementById("logLevel").value
        const javaOpts = document.getElementById("javaOpts").value

        const serverIP = document.getElementById("serverIP").value


        // build JSON object
        const jobRequest = {

            jobName: jobName,

            jar: {
                url: jarURL,
                mainClass: jarMainClass
            },

            resources: {
                executors: executors,
                coresPerExecutor: coresPerExecutor,
                memoryPerExecutorMB: memoryPerExecutorMB
            },

            data: {
                input: [dataInput],
                output: dataOutput
            },

            arguments: jarArguments.split(" "),

            environment: {
                LOG_LEVEL: logLevel,
                JAVA_OPTS: javaOpts
            },

            timeoutSeconds: timeoutSeconds,
            priority: priority
        }


        try {

            const response = await fetch(`http://${serverIP}/jobs/submit`, {

                method: "POST",

                headers: {
                    "Content-Type": "application/json"
                },

                body: JSON.stringify(jobRequest)

            })

            if (!response.ok) {

                throw new Error("Request failed")

            }

            const result = await response.json()

            console.log("job created:", result)

        } catch (err) {

            console.error("error:", err)

        }

    })

})