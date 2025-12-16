// Load the Go WASM support script
importScripts("/wasm_exec.js");

let isReady = false;

async function loadWasm() {
  const go = new self.Go();
  try {
    const result = await WebAssembly.instantiateStreaming(fetch("/viri.wasm"), go.importObject);
    go.run(result.instance);
    isReady = true;
    postMessage({ type: "ready" });
  } catch (err) {
    console.error("Worker failed to load WASM:", err);
    postMessage({ type: "error", content: "Failed to load Viri runtime." });
  }
}

loadWasm();

self.onmessage = (e) => {
  const { type, code } = e.data;

  if (type === "run") {
    if (!isReady) {
      postMessage({
        type: "error",
        content: "Runtime not ready. Please wait...",
      });
      return;
    }

    try {
      // runViri is exposed by the Go WASM on the global scope (self)
      // @ts-expect-error -- injected by Go WASM
      if (self.runViri) {
        // @ts-expect-error -- injected by Go WASM
        const raw = self.runViri(code);
        const result = JSON.parse(raw);
        postMessage({ type: "result", data: result });
      } else {
        postMessage({ type: "error", content: "Runtime function not found." });
      }
    } catch {
      postMessage({ type: "error", content: "Execution crashed." });
    }
  } else if (type === "reset") {
    // @ts-expect-error -- injected by Go WASM -- injected by Go WASM
    if (self.resetViri) {
      // @ts-expect-error -- injected by Go WASM
      self.resetViri();
    }
  }
};
