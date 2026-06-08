if (import.meta.main) {
  const listenPort = Deno.args[0]
  if (!listenPort) {
    throw new Error("missing listen port as positional argument")
  }

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
