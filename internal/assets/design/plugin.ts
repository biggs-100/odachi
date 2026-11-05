// odachi design: the push half of the canvas.
//
// The canvas writes one JSON file per message in the clone's store; this plugin
// reads the queue of the project it was opened in, sends each message to the
// session that asked for the canvas, and deletes the file. It is a thin bridge on
// purpose: the queue format is pinned by odachi's Go tests, and the canvas never
// depends on this plugin to save a comment.
import { Plugin } from "@opencode/plugin"
import { readdir, readFile, unlink } from "node:fs/promises"
import { join } from "node:path"

const QUEUE = [".git", "odachi", "design-queue"]
const ATTEMPTS = 3

export default Plugin.define({
  id: "odachi.design",
  async setup(ctx) {
    let running = false
    const failed = new Map<string, number>()

    const poll = async () => {
      if (running) return
      running = true
      try {
        const dir = join(ctx.location.directory, ...QUEUE)
        let entries
        try {
          entries = await readdir(dir)
        } catch {
          return
        }
        for (const name of entries) {
          if (!name.endsWith(".json")) continue
          const path = join(dir, name)
          try {
            const message = JSON.parse(await readFile(path, "utf8"))
            if (!message.sessionID || !message.text) {
              await unlink(path)
              continue
            }
            await ctx.session.prompt({ sessionID: message.sessionID, text: message.text, delivery: "queue" })
            await unlink(path)
            failed.delete(name)
          } catch (error) {
            const count = (failed.get(name) ?? 0) + 1
            failed.set(name, count)
            if (count >= ATTEMPTS) {
              // The comment stays in comments.json: the push is a convenience, and
              // a message that cannot be delivered is named once and dropped.
              await unlink(path).catch(() => {})
              failed.delete(name)
              console.error("[odachi-design] dropped a canvas message after three failed deliveries:", error)
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
