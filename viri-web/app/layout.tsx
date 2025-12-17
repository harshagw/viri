import type { Metadata } from "next";
import { Geist, Geist_Mono, JetBrains_Mono } from "next/font/google";
import { Navigation } from "@/components/navigation";
import "./globals.css";

const jetbrainsMono = JetBrains_Mono({ subsets: ["latin"], variable: "--font-sans" });

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "Viri",
  description: "Viri is a programming language that is designed to be easy to learn and use.",
};

import { ThemeProvider } from "@/app/providers/theme-provider";
import { PostHogProvider } from "@/app/providers/posthog-provider";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <PostHogProvider>
      <html lang="en" className={jetbrainsMono.variable} suppressHydrationWarning>
        <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
          <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
            <div className="min-h-screen flex flex-col bg-background">
              <Navigation />
              {children}
            </div>
          </ThemeProvider>
        </body>
      </html>
    </PostHogProvider>
  );
}
