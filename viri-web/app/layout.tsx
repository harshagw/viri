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
  description: "A simple, expressive programming language designed to be easy to learn and use.",
  metadataBase: new URL("https://harshagw.github.io/viri"),
  keywords: ["viri", "programming language", "interpreter", "compiler", "bytecode", "vm", "learning"],
  authors: [{ name: "Harsh Agarwal", url: "https://harshagw.github.io" }],
  creator: "Harsh Agarwal",
  openGraph: {
    title: "Viri",
    description: "A simple, expressive programming language designed to be easy to learn and use.",
    siteName: "Viri",
    type: "website",
    locale: "en_US",
    url: "https://harshagw.github.io/viri",
    images: [
      {
        url: "https://harshagw.github.io/viri/og-image.png",
        width: 1200,
        height: 630,
        alt: "Viri - A simple, expressive programming language",
      },
    ],
  },
  twitter: {
    card: "summary_large_image",
    site: "@harsh_agw",
    creator: "@harsh_agw",
    title: "Viri",
    description: "A simple, expressive programming language designed to be easy to learn and use.",
    images: {
      url: "https://harshagw.github.io/viri/og-image.png",
      alt: "Viri - A simple, expressive programming language",
    },
  },
  icons: {
    icon: "/favicon.svg",
    apple: "/favicon.svg",
  },
  other: {
    "linkedin:title": "Viri",
    "linkedin:description": "A simple, expressive programming language designed to be easy to learn and use.",
  },
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
