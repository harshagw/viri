"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { Loader2 } from "lucide-react";

import { Playground } from "@/components/playground";
import { getSnippet, DEFAULT_CODE } from "@/lib/snippets";

function PlaygroundContent() {
  const searchParams = useSearchParams();
  const snippetId = searchParams.get("snippet");
  const initialCode = (snippetId ? getSnippet(snippetId)?.code : null) ?? DEFAULT_CODE;

  // The dedicated page already fills the window, so there is nothing for a
  // full-screen control to do here.
  return <Playground initialCode={initialCode} className="h-[calc(100vh-75px)]" />;
}

export default function PlaygroundPage() {
  return (
    <Suspense
      fallback={
        <div className="flex h-screen items-center justify-center">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      }
    >
      <PlaygroundContent />
    </Suspense>
  );
}
