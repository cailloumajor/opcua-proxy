import { listenPort } from "./main.ts"

if (import.meta.main) {
    Error.stackTraceLimit = 0
    try {
        const response = await fetch(`http://127.0.0.1:${listenPort}/status`)
        if (!response.ok) {
            throw new Error(`Response status: ${response.status}`)
        }
    } catch (error) {
        console.error(error)
        Deno.exit(1)
    }
}
