export const listenPort = 3000

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
        _id: "integration-tests",
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

function addrToString({ transport, hostname, port }: Deno.NetAddr) {
    return `${hostname}:${port} (${transport})`
}

function onListen(addr: Deno.NetAddr) {
    console.log(`Listening on ${addrToString(addr)}`)
}

const route = new URLPattern({ pathname: "/:id" })

if (import.meta.main) {
    Deno.serve({ port: listenPort, onListen }, (req, info) => {
        const addrString = addrToString(info.remoteAddr)
        console.log(`Got a request from ${addrString}: ${req.method} ${req.url}`)

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
    })
}
