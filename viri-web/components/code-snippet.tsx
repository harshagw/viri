"use client";

import * as React from "react";
import { cn } from "@/lib/utils";
import Link from "next/link";
import { Button } from "./ui/button";

interface CodeSnippetProps {
  title: string;
  code: string;
  className?: string;
  id?: string;
}

export function CodeSnippet({ title, code, className, id }: CodeSnippetProps) {
  return (
    <div className={cn("flex flex-col relative group", className)}>
      <div className="flex justify-between items-end mb-2">
        <p className="text-xs text-muted-foreground font-mono">{title}</p>
        {id && (
          <Link href={`/playground?snippet=${id}`}>
            <Button variant="link" size="sm" className="uppercase opacity-0 group-hover:opacity-100 transition-opacity hover:underline">
              Try in playground -&gt;
            </Button>
          </Link>
        )}
      </div>
      <pre className="bg-muted p-5 text-sm border-l-2 border-primary flex-1 whitespace-pre-wrap wrap-break-word">
        <code className="font-mono">{code}</code>
      </pre>
    </div>
  );
}
