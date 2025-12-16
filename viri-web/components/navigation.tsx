"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { ThemeToggle } from "@/components/theme-toggle";

export function Navigation() {
  const pathname = usePathname();

  return (
    <nav className="border-b border-border">
      <div className="container mx-auto px-6 py-5 max-w-4xl">
        <div className="flex items-center justify-between">
          <Link href="/" className="font-mono font-bold text-primary hover:opacity-70">
            viri
          </Link>
          <div className="flex items-center gap-8">
            <div className="flex gap-8 text-sm">
              <Link href="/" className={pathname === "/" ? "text-foreground" : "text-muted-foreground hover:text-foreground"}>
                home
              </Link>
              <Link href="/grammar" className={pathname === "/grammar" ? "text-foreground" : "text-muted-foreground hover:text-foreground"}>
                grammar
              </Link>
            </div>
            <ThemeToggle />
          </div>
        </div>
      </div>
    </nav>
  );
}
