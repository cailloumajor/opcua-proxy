const config = [
  {
    _id: "novalue",
    serverUrl: "opc.tcp://opcua-server-first:50000",
    securityPolicy: "None",
    securityMode: "None",
    user: null,
    password: null,
    tags: [
      {
        type: "tag",
        name: "writeOnly",
        namespaceUri: "http://microsoft.com/Opc/OpcPlc/ReferenceTest",
        nodeIdentifier: "AccessRights_AccessAll_WO",
      },
    ],
  },
  {
    _id: "moving",
    serverUrl: "opc.tcp://opcua-server-second:50000",
    securityPolicy: "Basic256Sha256",
    securityMode: "SignAndEncrypt",
    user: "user1",
    password: "password",
    tags: [
      {
        type: "container",
        namespaceUri: "http://opcfoundation.org/UA/",
        nodeIdentifier: 2256,
      },
      {
        type: "container",
        namespaceUri: "http://microsoft.com/Opc/OpcPlc/",
        nodeIdentifier: "Basic",
      },
      {
        type: "tag",
        name: "slowNumberOfUpdates",
        namespaceUri: "http://microsoft.com/Opc/OpcPlc/",
        nodeIdentifier: "SlowNumberOfUpdates",
      },
    ],
  },
]

const route = new URLPattern({ pathname: "/:id" })

export default {
  fetch(req, info) {
    if (info.remoteAddr.transport === "tcp") {
      const { hostname, port } = info.remoteAddr
      console.log(`Got a request from ${hostname}:${port}: ${req.method} ${req.url}`)
    }

    if (req.method !== "GET") {
      return new Response("Method Not Allowed", { status: 405 })
    }

    const id = route.exec(req.url)?.pathname.groups.id

    switch (id) {
      case "config":
        return Response.json(config)
      case "status":
        return new Response(null, { status: 204 })
      default:
        return new Response("Not Found", { status: 404 })
    }
  },
} satisfies Deno.ServeDefaultExport
