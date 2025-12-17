import { basePath } from "@/lib/utils";
import { useEffect, useState, useCallback, useRef } from "react";

interface ViriResponse {
  result: string;
  output: string;
  errors: string[];
  warnings: string[];
}

interface UseViriReturn {
  isReady: boolean;
  isWasmSupported: boolean;
  run: (code: string) => void;
  reset: () => void;
  clear: () => void;
  result: ViriResponse | null;
  isRunning: boolean;
  error: string | null;
}

const TIMEOUT_MS = 10000; // 10s timeout
const MAX_INPUT_SIZE = 10000; // 10KB

export function useViriPlayground(): UseViriReturn {
  const [isReady, setIsReady] = useState(false);
  const [isWasmSupported, setIsWasmSupported] = useState(true);
  const [isRunning, setIsRunning] = useState(false);
  const [result, setResult] = useState<ViriResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const workerRef = useRef<Worker | null>(null);
  const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (typeof WebAssembly !== "object") {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setIsWasmSupported(false);
    }
  }, []);

  const initWorker = useCallback(() => {
    if (typeof window === "undefined") return;

    // Check WASM support
    if (typeof WebAssembly !== "object") {
      return;
    }

    if (workerRef.current) {
      workerRef.current.terminate();
    }

    const worker = new Worker(new URL("./viri.worker.js", import.meta.url), { type: "classic" });

    workerRef.current = worker;

    worker.postMessage({ type: "init", basePath: basePath });

    worker.onmessage = (e) => {
      const { type, data, content } = e.data;

      if (type === "ready") {
        setIsReady(true);
        return;
      }

      if (type === "result") {
        if (timeoutRef.current) clearTimeout(timeoutRef.current);
        setResult(data);
        setIsRunning(false);
        return;
      }

      if (type === "error") {
        if (timeoutRef.current) clearTimeout(timeoutRef.current);
        setError(content);
        setIsRunning(false);
      }
    };

    worker.onerror = (e) => {
      console.error("Worker crashed", e);
      setError("Runtime worker crashed.");
      setIsRunning(false);
    };
  }, []);

  useEffect(() => {
    initWorker();
    return () => {
      // Cleanup on unmount
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      if (workerRef.current) workerRef.current.terminate();
    };
  }, [initWorker]);

  const run = useCallback(
    (code: string) => {
      if (!workerRef.current || !isReady) return;

      if (code.length > MAX_INPUT_SIZE) {
        setError(`Input too large. Maximum size is ${MAX_INPUT_SIZE} characters.`);
        return;
      }

      setIsRunning(true);
      setResult(null);
      setError(null);

      workerRef.current.postMessage({ type: "reset" });
      workerRef.current.postMessage({ type: "run", code });

      // Set timeout
      if (timeoutRef.current) clearTimeout(timeoutRef.current);
      timeoutRef.current = setTimeout(() => {
        setError(`Execution timed out after ${TIMEOUT_MS / 1000}s.`);
        setIsRunning(false);

        // Terminate and restart worker to clear any stuck state
        if (workerRef.current) {
          workerRef.current.terminate();
          setIsReady(false);
          initWorker();
        }
      }, TIMEOUT_MS);
    },
    [isReady, initWorker]
  );

  const reset = useCallback(() => {
    if (workerRef.current) {
      workerRef.current.postMessage({ type: "reset" });
    }
  }, []);

  const clear = useCallback(() => {
    setResult(null);
    setError(null);
  }, []);

  return {
    isReady,
    isWasmSupported,
    run,
    reset,
    clear,
    result,
    isRunning,
    error,
  };
}
