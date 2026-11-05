// odachi design: the push half of the canvas.
//
// The canvas writes one JSON file per message in the clone's store; this plugin
// reads the queue of the project it was opened in, sends each message to the
// session that asked for the canvas, and deletes the file. It claims a message by
// renaming it before delivery, so a retry after a failed unlink cannot send the
// same comment twice; three failed deliveries leave a `.failed` file and a log
// line instead of disappearing.
import { Plugin } from "@opencode/plugin"
import { readdir, readFile, rename, unlink } from "node:fs/promises"
import { join } from "node:path"

const QUEUE = [".git", "odachi", "design-queue"]
const ATTEMPTS = 3
const CLAIM = ".sending"

export default Plugin.define({
  id: "odachi.design",
  async setup(ctx) {
    const dir = join(ctx.location.directory, ...QUEUE)
    let running = false
    const failed = new Map<string, number>()

    // A process that died mid-delivery leaves a claim; the next start returns it
    // to the queue, because no delivery of this instance is in flight yet.
    try {
      for (const name of await readdir(dir)) {
        if (!name.endsWith(CLAIM)) continue
        await rename(join(dir, name), join(dir, name.slice(0, -CLAIM.length))).catch(() => {})
      }
    } catch {
      // The project has no queue yet; the first comment creates it.
    }

    const poll = async () => {
      if (running) return
      running = true
      try {
        let entries
        try {
          entries = await readdir(dir)
        } catch {
          return
        }
        for (const name of entries) {
          if (!name.endsWith(".json")) continue
          const path = join(dir, name)
          const claim = join(dir, name + CLAIM)
          try {
            await rename(path, claim)
          } catch {
            // Another poll or another instance owns it.
            continue
          }
          try {
            const message = JSON.parse(await readFile(claim, "utf8"))
            if (!message.sessionID || !message.text) {
              await unlink(claim)
              continue
            }
            await ctx.session.prompt({ sessionID: message.sessionID, text: message.text, delivery: "queue" })
            await unlink(claim)
            failed.delete(name)
          } catch (error) {
            const count = (failed.get(name) ?? 0) + 1
            failed.set(name, count)
            if (count >= ATTEMPTS) {
              // The comment stays in comments.json: the push is a convenience, and
              // a message that cannot be delivered is named and kept, never silent.
              failed.delete(name)
              await rename(claim, join(dir, name.replace(/\.json$/, ".failed"))).catch(() => {})
              console.error("[odachi-design] dropped a canvas message after three failed deliveries:", error)
            } else {
              await rename(claim, path).catch(() => {})
            }
          }
        }
      } finally {
        running = false
      }
    }

    const timer = setInterval(() => void poll(), 750)
    void poll()
    return () => clearInterval(timer)
  },
})
