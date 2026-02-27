import { useState, useCallback } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useShockClient } from "./provider.js";
import { shockKeys } from "./hooks.js";
import { smartUpload } from "../upload.js";
import type { ShockNode, SmartUploadOptions, UploadProgress } from "../types.js";

export interface UseUploadInput {
  file: Blob;
  fileName?: string;
  attributes?: Record<string, unknown>;
}

export interface UseUploadOptions {
  chunkSize?: number;
}

export interface UseUploadReturn<TAttr = unknown> {
  upload: (input: UseUploadInput) => void;
  progress: UploadProgress | null;
  isUploading: boolean;
  error: Error | null;
  node: ShockNode<TAttr> | null;
  reset: () => void;
}

export function useUpload<TAttr = unknown>(
  options?: UseUploadOptions
): UseUploadReturn<TAttr> {
  const client = useShockClient();
  const qc = useQueryClient();
  const [progress, setProgress] = useState<UploadProgress | null>(null);

  const mutation = useMutation<ShockNode<TAttr>, Error, UseUploadInput>({
    mutationFn: (input: UseUploadInput) => {
      setProgress({ loaded: 0, total: 0, percent: 0 });

      const uploadOptions: SmartUploadOptions = {
        fileName: input.fileName,
        attributes: input.attributes,
        chunkSize: options?.chunkSize,
        onProgress: setProgress,
      };

      return smartUpload<TAttr>(client, input.file, uploadOptions);
    },
    onSuccess: (node) => {
      // Warm the single-node cache
      qc.setQueryData(shockKeys.node(node.id), node);
      // Invalidate node lists
      qc.invalidateQueries({ queryKey: ["shock", "nodes"] });
    },
    onError: () => {
      setProgress(null);
    },
  });

  const reset = useCallback(() => {
    setProgress(null);
    mutation.reset();
  }, [mutation]);

  return {
    upload: mutation.mutate,
    progress,
    isUploading: mutation.isPending,
    error: mutation.error,
    node: mutation.data ?? null,
    reset,
  };
}
