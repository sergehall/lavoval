export type RuntimePreset = {
  entrypoint: string;
  title: string;
  description: string;
  inputLabel: string;
  inputPlaceholder: string;
  outputHint: string;
};

export const runtimePresets: Record<string, RuntimePreset> = {
  echo: {
    entrypoint: 'echo',
    title: 'Echo',
    description: 'Mirrors the provided input so you can verify the runtime pipeline end to end.',
    inputLabel: 'Input payload',
    inputPlaceholder: 'Write any text here to see the exact payload echoed back.',
    outputHint: 'Returns the input exactly as it came in.',
  },
  'text-summary-mock': {
    entrypoint: 'text-summary-mock',
    title: 'Text Summary Mock',
    description:
      'Creates a shortened summary from the supplied text using deterministic mock rules.',
    inputLabel: 'Source text',
    inputPlaceholder: 'Paste a long paragraph, notes, or article excerpt to generate a shorter summary.',
    outputHint: 'Returns a shortened `result` string.',
  },
  'keyword-extract-mock': {
    entrypoint: 'keyword-extract-mock',
    title: 'Keyword Extract Mock',
    description:
      'Extracts the strongest repeated keywords from the supplied text using mock heuristics.',
    inputLabel: 'Source text',
    inputPlaceholder:
      'Paste notes, a transcript, or a paragraph to extract repeated keywords and topic signals.',
    outputHint: 'Returns a `keywords` array.',
  },
};

export function getRuntimePreset(entrypoint: string): RuntimePreset {
  return (
    runtimePresets[entrypoint] ?? {
      entrypoint,
      title: 'Custom Runtime',
      description:
        'This skill uses a custom runtime entrypoint. For the current milestone, provide text input unless the executor expects something else.',
      inputLabel: 'Input text',
      inputPlaceholder: 'Paste the input you want to send into this skill runtime.',
      outputHint: 'Output depends on the selected entrypoint.',
    }
  );
}
