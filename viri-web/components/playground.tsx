"use client";

import { useState, useRef, useEffect } from "react";
import { Play, RotateCcw, Terminal, AlertCircle, AlertTriangle, Loader2, Maximize2, Minimize2, Download } from "lucide-react";
import { useDownload } from "@/hooks/use-download";
import { useViriPlayground } from "@/hooks/use-viri-playground";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { ScrollArea } from "@/components/ui/scroll-area";
import { cn } from "@/lib/utils";
import { DEFAULT_CODE } from "@/lib/snippets";

import Editor from "react-simple-code-editor";
import { highlight, languages } from "prismjs";
import { registerViri } from "@/lib/prism-viri";
import "@/app/prism-viri.css";

// Register the language immediately so the first paint is highlighted.
registerViri();

interface PlaygroundProps {
  /** Code to start with. Defaults to the hello-world snippet. */
  initialCode?: string;
  /** Height of the pane when not full screen. */
  className?: string;
  /** Show the control that expands the playground to fill the window. */
  allowFullscreen?: boolean;
}

export function Playground({ initialCode, className, allowFullscreen = false }: PlaygroundProps) {
  const { isReady, isWasmSupported, run, clear, result, isRunning, error } = useViriPlayground();
  const { download } = useDownload();

  const [code, setCode] = useState(initialCode ?? DEFAULT_CODE);
  const [isFullscreen, setIsFullscreen] = useState(false);

  const editorContainerRef = useRef<HTMLDivElement>(null);
  const lineNumbersRef = useRef<HTMLDivElement>(null);

  const lineCount = code.split("\n").length;
  const lineNumbers = Array.from({ length: lineCount }, (_, i) => i + 1);

  const handleScroll = () => {
    if (editorContainerRef.current && lineNumbersRef.current) {
      lineNumbersRef.current.scrollTop = editorContainerRef.current.scrollTop;
    }
  };

  useEffect(() => {
    registerViri();
  }, []);

  // Escape leaves full screen, which is what every full-screen surface does.
  useEffect(() => {
    if (!isFullscreen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") setIsFullscreen(false);
    };
    window.addEventListener("keydown", onKey);
    // Stop the page behind from scrolling while the overlay is up.
    const previousOverflow = document.body.style.overflow;
    document.body.style.overflow = "hidden";
    return () => {
      window.removeEventListener("keydown", onKey);
      document.body.style.overflow = previousOverflow;
    };
  }, [isFullscreen]);

  if (!isWasmSupported) {
    return (
      <div className="flex h-[50vh] flex-col items-center justify-center gap-4 text-center p-6">
        <Alert variant="destructive" className="max-w-md">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>WebAssembly Not Supported</AlertTitle>
          <AlertDescription>
            Your browser does not support WebAssembly, which is required to run the Viri Playground. Please upgrade your browser or check your settings.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  return (
    <div
      className={cn(
        "flex flex-col md:flex-row overflow-hidden font-sans bg-background text-foreground",
        isFullscreen ? "fixed inset-0 z-50 h-screen" : className,
      )}
    >
      {/* Input Pane */}
      <section className="flex-1 flex flex-col border-b md:border-b-0 md:border-r border-border bg-card relative group min-w-0 min-h-0">
        <div className="shrink-0 bg-muted/30 border-b border-border px-4 py-2 flex justify-between items-center select-none h-[45px]">
          <span className="text-xs font-mono text-muted-foreground uppercase tracking-wider truncate">Input</span>

          <div className="flex items-center gap-2 sm:gap-3 shrink-0">
            {!isReady && !error && <Badge variant="secondary" className={cn("text-yellow-500 bg-yellow-500/10")}>Loading...</Badge>}
            {(allowFullscreen || isFullscreen) && (
              <Button
                onClick={() => setIsFullscreen((v) => !v)}
                variant="secondary"
                size="sm"
                aria-label={isFullscreen ? "Exit full screen" : "Enter full screen"}
              >
                {isFullscreen ? <Minimize2 className="w-3 h-3" /> : <Maximize2 className="w-3 h-3" />}
                <span className="hidden sm:inline">{isFullscreen ? "Exit" : "Full screen"}</span>
              </Button>
            )}
            <Button
              onClick={() => {
                setCode("");
                clear();
              }}
              variant="secondary"
              size="sm"
            >
              <RotateCcw className="w-3 h-3" />
              <span className="hidden sm:inline">Clear</span>
            </Button>
            <Button onClick={() => run(code)} disabled={!isReady || isRunning} size="sm">
              {isRunning ? <Loader2 className="w-3 h-3 animate-spin" /> : <Play className="w-3 h-3 fill-current" />}
              Run
            </Button>
          </div>
        </div>

        <div className="flex-1 flex relative min-h-0">
          {/* Line Numbers */}
          <div
            ref={lineNumbersRef}
            className="w-12 shrink-0 bg-muted/20 border-r border-border text-right py-4 pr-3 text-sm font-mono text-muted-foreground/50 select-none overflow-hidden"
            style={{ paddingTop: 16 }}
          >
            {lineNumbers.map((n) => (
              <div key={n} className="leading-6 h-[24px]">
                {n}
              </div>
            ))}
          </div>

          {/* Code Editor */}
          <div ref={editorContainerRef} className="flex-1 overflow-auto bg-muted/20 cursor-text" onScroll={handleScroll}>
            <Editor
              value={code}
              onValueChange={setCode}
              highlight={(code) => highlight(code, languages.viri, "viri")}
              padding={16}
              className="font-mono min-h-full"
              textareaClassName="focus:outline-none"
              style={{
                fontFamily: "monospace",
                fontSize: 14,
                backgroundColor: "transparent",
                minHeight: "100%",
                lineHeight: "24px",
              }}
            />
          </div>
        </div>
      </section>

      {/* Output Pane */}
      <section className="flex-1 flex flex-col bg-muted/10 relative min-w-0 min-h-0">
        <div className="shrink-0 bg-muted/50 border-b border-border px-4 py-3 text-xs font-mono text-muted-foreground uppercase tracking-wider flex items-center gap-2 select-none h-[45px]">
          <Terminal className="w-3 h-3" />
          <span>Output</span>
        </div>

        <div className="flex-1 min-h-0 relative">
          <ScrollArea className="h-full w-full bg-transparent">
            <div className="p-6 font-mono text-sm space-y-3">
              {error && (
                <Alert variant="destructive">
                  <AlertCircle className="h-4 w-4" />
                  <AlertTitle>System Error</AlertTitle>
                  <AlertDescription>{error}</AlertDescription>
                </Alert>
              )}

              {!result && !error && !isRunning && <div className="text-muted-foreground italic opacity-50">Results will appear here...</div>}

              {isRunning && !result && <div className="text-muted-foreground animate-pulse">Running...</div>}

              {result && (
                <>
                  {/* Type and runtime errors */}
                  {result.errors && result.errors.length > 0 && (
                    <div className="space-y-1">
                      {result.errors.map((err: string, i: number) => (
                        <Alert key={i} variant="destructive">
                          <AlertCircle className="h-4 w-4" />
                          <AlertDescription>{err}</AlertDescription>
                        </Alert>
                      ))}
                    </div>
                  )}

                  {result.warnings && result.warnings.length > 0 && (
                    <div className="space-y-1">
                      {result.warnings.map((warn: string, i: number) => (
                        <Alert key={i} variant="warning">
                          <AlertTriangle className="h-4 w-4" />
                          <AlertDescription>Warning: {warn}</AlertDescription>
                        </Alert>
                      ))}
                    </div>
                  )}

                  {result.output && (
                    <div className="whitespace-pre-wrap text-foreground">
                      {result.output.length > 2000 ? (
                        <>
                          {result.output.slice(0, 2000)}
                          <div className="flex items-center gap-3 mt-4 pt-3 border-t border-border">
                            <div className="text-muted-foreground italic text-xs">...Output truncated (too long)...</div>
                            <Button variant="outline" size="sm" onClick={() => download("output.txt", result.output)}>
                              <Download className="w-3 h-3 mr-2" />
                              Download Full Output
                            </Button>
                          </div>
                        </>
                      ) : (
                        result.output.trimEnd()
                      )}
                    </div>
                  )}

                  {result.result && <div className="text-primary font-bold mt-2 border-l-2 border-primary pl-3 py-1">{result.result}</div>}
                </>
              )}
            </div>
          </ScrollArea>
        </div>
      </section>
    </div>
  );
}
