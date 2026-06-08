import { assert, assertEquals } from "@std/assert"
import { Centrifuge } from "centrifuge"
import * as z from "zod"

const movingChannelValues = z.object({
  StartTime: z.string(),
  CurrentTime: z.iso.datetime(),
  State: z.int().nonnegative().lte(7),
  BuildInfo: z.looseObject({}),
  SecondsTillShutdown: z.int().nonnegative(),
  slowNumberOfUpdates: z.int(),
  RandomSignedInt32: z.int32(),
  StepUp: z.int().nonnegative(),
  RandomUnsignedInt32: z.int32().nonnegative(),
  AlternatingBoolean: z.boolean(),
})

const movingChannelTimestamps = z.record(movingChannelValues.keyof(), z.iso.datetime())

const movingChannelInitialData = z.object({
  val: movingChannelValues,
  ts: movingChannelTimestamps,
})

const movingChannelPublicationData = z.object({
  val: movingChannelValues.partial(),
  ts: z.partialRecord(movingChannelTimestamps.keyType, movingChannelTimestamps.valueType),
})

let noDataSubscribed = false
let movingSubscribed = false
let publicationsReceived = 0

const centrifuge = new Centrifuge("ws://centrifugo:8000/connection/websocket", {
  debug: true,
})

const noDataSub = centrifuge.newSubscription("integration.tests:nodata")

noDataSub.on("subscribed", ({ channel, data }) => {
  console.log(
    "got subscribed event for `%s` channel, with data: %O",
    channel,
    data,
  )
  noDataSubscribed = true
  console.log("testing no-data channel data object")
  assertEquals(data, undefined)
})

noDataSub.on("publication", () => {
  throw new Error("unexpected publication on no-data channel")
})

noDataSub.subscribe()

const sub = centrifuge.newSubscription("integration.tests:moving")

sub.on("subscribed", ({ channel, data }) => {
  console.log(
    "got subscribed event for `%s` channel, with data: %O",
    channel,
    data,
  )

  movingSubscribed = true

  movingChannelInitialData.parse(data)
})

sub.on("publication", ({ data }) => {
  console.log("got publication with data: %O", data)

  publicationsReceived += 1

  const parsed = movingChannelPublicationData.parse(data)

  assertEquals(Object.keys(parsed.val), Object.keys(parsed.ts))
})

sub.subscribe()

centrifuge.connect()

const publicationsTimeout = setTimeout(() => {
  throw new Error("timeout waiting for publications array to populate")
}, 10000)

await new Promise((resolve) => {
  const interval = setInterval(() => {
    if (publicationsReceived >= 6) {
      clearInterval(interval)
      resolve(null)
    }
  }, 500)
})

centrifuge.disconnect()
clearTimeout(publicationsTimeout)

assert(noDataSubscribed, "`no-data` channel has not been subscribed")

assert(movingSubscribed, "`moving` channel has not been subscribed")
