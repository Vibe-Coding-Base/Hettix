// Shared encoders/decoders used by the Decoder tool and the editor's in-place
// decode context menu. Each function throws on invalid input so callers can
// surface the error.

const utf8 = new TextEncoder();
const utf8dec = new TextDecoder();

export function base64Encode(s: string): string {
  const bytes = utf8.encode(s);
  let bin = "";
  bytes.forEach((b) => (bin += String.fromCharCode(b)));
  return btoa(bin);
}

export function base64Decode(s: string): string {
  const bin = atob(s.trim());
  const bytes = Uint8Array.from(bin, (c) => c.charCodeAt(0));
  return utf8dec.decode(bytes);
}

export function hexEncode(s: string): string {
  return Array.from(utf8.encode(s))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

export function hexDecode(s: string): string {
  const clean = s.replace(/\s+/g, "");
  if (clean.length % 2 !== 0) {
    throw new Error("hex input must have an even number of digits");
  }
  const bytes = new Uint8Array(clean.length / 2);
  for (let i = 0; i < bytes.length; i++) {
    bytes[i] = parseInt(clean.substr(i * 2, 2), 16);
  }
  return utf8dec.decode(bytes);
}

export function htmlEncode(s: string): string {
  return s.replace(/[&<>"']/g, (c) => `&#${c.charCodeAt(0)};`);
}

export function htmlDecode(s: string): string {
  const el = document.createElement("textarea");
  el.innerHTML = s;
  return el.value;
}

export interface Codec {
  label: string;
  encode: (s: string) => string;
  decode: (s: string) => string;
}

export const CODECS: Codec[] = [
  { label: "Base64", encode: base64Encode, decode: base64Decode },
  { label: "URL", encode: encodeURIComponent, decode: decodeURIComponent },
  { label: "HTML", encode: htmlEncode, decode: htmlDecode },
  { label: "Hex", encode: hexEncode, decode: hexDecode },
];
