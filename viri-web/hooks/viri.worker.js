let isReady = false;
let basePath = "";

self.onmessage = async (e) => {
  if (e.data.type === "init") {
    basePath = e.data.basePath || "";
    await loadWasm();
    return;
  }

  if (e.data.type === "run") {
    if (!isReady) {
      postMessage({ type: "error", content: "Runtime not ready." });
      return;
    }

    try {
      const raw = self.runViri(e.data.code);
      postMessage({ type: "result", data: JSON.parse(raw) });
    } catch {
      postMessage({ type: "error", content: "Execution crashed." });
    }
  }

  if (e.data.type === "reset") {
    self.resetViri?.();
  }
};

async function loadWasm() {
  try {
    const origin = self.location.origin;
    const wasmExecUrl = `${origin}${basePath}/wasm_exec.js`;
    const wasmUrl = `${origin}${basePath}/viri.wasm`;

    importScripts(wasmExecUrl);

    const go = new self.Go();
    const res = await fetch(wasmUrl);
    const bytes = await res.arrayBuffer();

    const { instance } = await WebAssembly.instantiate(bytes, go.importObject);

    go.run(instance);

    isReady = true;
    postMessage({ type: "ready" });
  } catch (err) {
    console.error(err);
    postMessage({
      type: "error",
      content: "Failed to initialize Viri runtime.",
    });
  }
}
