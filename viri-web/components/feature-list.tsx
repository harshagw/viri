"use client";

import * as React from "react";
import { cn } from "@/lib/utils";

interface Feature {
  title: string;
  description: string;
}

interface FeatureListProps {
  features: Feature[];
  className?: string;
}

export function FeatureList({ features, className }: FeatureListProps) {
  return (
    <div className={cn("grid md:grid-cols-2 gap-x-16 gap-y-6 text-sm", className)}>
      {features.map((feature, index) => (
        <div key={index} className="flex gap-4">
          <span className="text-primary">○</span>
          <div>
            <span className="font-medium">{feature.title}</span>
            <span className="text-muted-foreground"> — {feature.description}</span>
          </div>
        </div>
      ))}
    </div>
  );
}
