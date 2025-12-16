"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

interface CodeSnippetProps {
  title: string;
  code: string;
  className?: string;
}

export function CodeSnippet({ title, code, className }: CodeSnippetProps) {
  return (
    <div className={cn("flex flex-col", className)}>
      <p className="text-xs text-muted-foreground mb-2 font-mono">{title}</p>
      <pre className="bg-muted p-5 text-sm border-l-2 border-primary overflow-x-auto flex-1">
        <code className="font-mono">{code}</code>
      </pre>
    </div>
  );
}
