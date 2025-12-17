"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

import { ThemeToggle } from "@/components/theme-toggle";

import { Menu } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger, SheetTitle } from "@/components/ui/sheet";
import { useState } from "react";

export function Navigation() {
  const pathname = usePathname();
  const [isOpen, setIsOpen] = useState(false);

  const links = [
    { href: "/", label: "home" },
    { href: "/grammar", label: "grammar" },
    { href: "/playground", label: "playground" },
  ];

  return (
    <nav className="border-b border-border sticky top-0 bg-background z-50">
      <div className="container mx-auto px-6 py-4 max-w-4xl">
        <div className="flex items-center justify-between">
          <Link href="/" className="font-mono font-bold text-primary hover:opacity-70 text-lg">
            viri
          </Link>

          {/* Desktop Navigation */}
          <div className="hidden md:flex items-center gap-8">
            <div className="flex gap-8 text-sm font-medium">
              {links.map((link) => (
                <Link
                  key={link.href}
                  href={link.href}
                  className={pathname === link.href ? "text-foreground" : "text-muted-foreground hover:text-foreground transition-colors"}
                >
                  {link.label}
                </Link>
              ))}
            </div>
            <ThemeToggle />
          </div>

          {/* Mobile Navigation */}
          <div className="flex items-center gap-4 md:hidden">
            <ThemeToggle />
            <Sheet open={isOpen} onOpenChange={setIsOpen}>
              <SheetTrigger
                render={
                  <Button variant="ghost" size="icon" className="-mr-2">
                    <Menu className="h-5 w-5" />
                    <span className="sr-only">Toggle menu</span>
                  </Button>
                }
              />
              <SheetContent side="right" className="p-6">
                <SheetTitle className="font-mono font-bold text-primary mb-6 text-lg">viri</SheetTitle>
                <div className="flex flex-col gap-6">
                  {links.map((link) => (
                    <Link
                      key={link.href}
                      href={link.href}
                      onClick={() => setIsOpen(false)}
                      className={
                        pathname === link.href ? "text-foreground font-medium text-lg" : "text-muted-foreground hover:text-foreground transition-colors text-lg"
                      }
                    >
                      {link.label}
                    </Link>
                  ))}
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </nav>
  );
}
