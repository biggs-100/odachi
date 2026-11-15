# UI / desktop verification

Three levels, standardized. Run them in this order, the ladder `drive` names:
focused, then the suite, then build and lint, then a short runtime probe. Never
skip a level because a lower one passed.

The contract these three levels satisfy: drivable, observable, inventoried, one
command, and a capture packet whose screenshots are asserted. A ladder without that
contract produces screenshots nobody asserts on, which is not verification.

| Level | Tooling | Purpose | Cost | Does NOT cover |
| --- | --- | --- | --- | --- |
| 1 | `vitest` + `@testing-library/react` + `jsdom` | Component logic, hooks, stores, rendering | Seconds; no browser, no Rust | Real IPC, Rust behavior, the desktop shell |
| 2 | Playwright against the web build (`vite preview`) | Full frontend flows against a served build | Browser; build + preview server | `invoke()`/IPC, Rust, the packaged desktop app |
| 3 | WebdriverIO + `@wdio/tauri-service` | Real desktop e2e on the built binary | Heaviest; Rust toolchain + built debug binary | — (this is the only level that sees the desktop app) |

Neither Level 1 nor Level 2 sees the packaged desktop app: they run JS against jsdom or a
web server, with no Rust process and no real IPC. That gap is exactly why Level 3 exists.

## Level 1 — component tests

- Tooling: `vitest`, `@testing-library/react`, `jsdom`.
- Script shape (replaces the scaffold's stub):

  ```json
  {
    "scripts": {
      "test": "vitest run",
      "test:watch": "vitest"
    }
  }
  ```

- The reusable scaffold `project-template/frontend/package.json` currently ships no test
  layer (`"test": "echo \"Error: no test specified\" && exit 1"`). Fixing that stub is
  part of scaffolding.
- Does NOT cover: anything that crosses IPC or requires Rust.

## Level 2 — web flow tests

- Tooling: Playwright against the web build served by `vite preview`.
- Script shape:

  ```json
  {
    "scripts": {
      "preview": "vite preview",
      "test:e2e:web": "playwright test"
    }
  }
  ```

- Does NOT cover: `invoke()`, Rust behavior, or the desktop window.

## Level 3 — desktop e2e

- Tooling: WebdriverIO + `@wdio/tauri-service`, maintained under the WebdriverIO project
  and the recommended path. It runs an embedded WebDriver server inside the app and adds
  `browser.tauri.execute()`, command (IPC) mocking, frontend and backend log capture, and
  multiremote.
- Script shape:

  ```json
  {
    "scripts": {
      "test:e2e:desktop": "wdio run wdio.conf.ts"
    }
  }
  ```

- `wdio.conf.ts` fragment:

  ```ts
  // wdio.conf.ts (fragment)
  export const config = {
    services: [
      ['@wdio/tauri-service', {
        autoInstallTauriDriver: true,
        autoDownloadEdgeDriver: true,
        captureBackendLogs: true,
        captureFrontendLogs: true,
      }],
    ],
  };
  ```

  UNVERIFIED: the surrounding `runner` / `specs` / capability keys are version-specific —
  copy them from the service's own example config rather than from memory.

### Platform matrix

| Platform | Driver | Notes |
| --- | --- | --- |
| Windows | auto-managed `msedgedriver.exe` | the service handles Edge WebDriver version matching |
| Linux | manual `WebKitWebDriver` install | not auto-managed |
| macOS | `embedded` / `crabnebula` driver providers | works through those providers |

Driving `tauri-driver` manually WITHOUT the service is Windows and Linux only.

### Requirements

- The Rust toolchain, to build the app and `tauri-driver`.
- On Windows also MSVC Build Tools.
- Real desktop e2e needs a built debug binary — the service drives the app you built.

## Sources

- <https://v2.tauri.app/develop/tests/webdriver/>
- <https://webdriver.io/docs/desktop-testing/tauri/platform-support/>
