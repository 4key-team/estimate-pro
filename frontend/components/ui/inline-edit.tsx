// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

"use client";

import { useState, useRef, useEffect } from "react";
import { Input } from "@/components/ui/input";
import { Pencil } from "lucide-react";

interface InlineEditProps {
  value: string;
  onSave: (newValue: string) => void;
  className?: string;
  inputClassName?: string;
  disabled?: boolean;
}

export function InlineEdit({
  value,
  onSave,
  className = "",
  inputClassName = "",
  disabled = false,
}: InlineEditProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [draft, setDraft] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    setDraft(value);
  }, [value]);

  useEffect(() => {
    if (isEditing) {
      inputRef.current?.focus();
      inputRef.current?.select();
    }
  }, [isEditing]);

  const commit = () => {
    const trimmed = draft.trim();
    if (trimmed && trimmed !== value) {
      onSave(trimmed);
    } else {
      setDraft(value);
    }
    setIsEditing(false);
  };

  const cancel = () => {
    setDraft(value);
    setIsEditing(false);
  };

  if (isEditing) {
    return (
      <Input
        ref={inputRef}
        value={draft}
        onChange={(e) => setDraft(e.target.value)}
        onBlur={commit}
        onKeyDown={(e) => {
          if (e.key === "Enter") commit();
          if (e.key === "Escape") cancel();
        }}
        className={inputClassName}
      />
    );
  }

  return (
    <button
      type="button"
      onClick={() => !disabled && setIsEditing(true)}
      className={`group inline-flex items-center gap-1.5 text-left ${disabled ? "cursor-default" : "cursor-pointer"} ${className}`}
    >
      <span>{value}</span>
      {!disabled && (
        <Pencil className="h-3.5 w-3.5 text-muted-foreground opacity-0 group-hover:opacity-100 transition-opacity" />
      )}
    </button>
  );
}
