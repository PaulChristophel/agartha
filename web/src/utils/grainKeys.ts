const extractSegments = (key: string, pattern: RegExp): string[] => {
  const matches = Array.from(key.matchAll(pattern));
  return matches.map((match) => match[1]);
};

export const toColonNotation = (rawKey: string): string => {
  if (!rawKey) {
    return rawKey;
  }

  const doubleQuoted = extractSegments(rawKey, /"([^"]+)"/g);
  if (doubleQuoted.length > 0) {
    return doubleQuoted.join(':');
  }

  const singleQuoted = extractSegments(rawKey, /'([^']+)'/g);
  if (singleQuoted.length > 0) {
    return singleQuoted.join(':');
  }

  return rawKey;
};

export const toJSONPathNotation = (colonKey: string): string => {
  if (!colonKey) {
    return colonKey;
  }

  // Existing dot-based keys pass through unchanged
  if (colonKey.includes('.') && !colonKey.includes(':')) {
    return colonKey;
  }

  return colonKey.split(':').join('.');
};
