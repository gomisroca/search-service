"""
Tokenizer - direct translation of internal/index/tokenizer.go.
 
The logic is identical: lowercase, replace non-alphanumeric characters with
spaces, split, skip stop words and short tokens. The only Python-specific
note is that we use str.isalnum() and str.lower() instead of Go's
unicode.IsLetter/unicode.IsDigit/unicode.ToLower, but they're equivalent.
"""
 
from __future__ import annotations
 
STOP_WORDS = frozenset({
    "a", "an", "and", "are", "as", "at", "be", "been", "but", "by",
    "do", "for", "from", "get", "has", "have", "he", "her", "his", "how",
    "i", "if", "in", "is", "it", "its", "just", "me", "my", "no",
    "not", "of", "on", "or", "our", "out", "she", "so", "than", "that",
    "the", "their", "them", "then", "there", "they", "this", "to", "up",
    "us", "was", "we", "were", "what", "when", "which", "who", "will",
    "with", "you", "your",
})
 
 
def tokenize(text: str, min_length: int) -> list[str]:
    """
    Lowercase, strip punctuation, remove stop words and short tokens.
    Identical output to the Go version's Tokenize() for the same input.
    """
    # Replace non-alphanumeric characters with spaces so "hello,world"
    # becomes "hello world" rather than "helloworld".
    cleaned = "".join(c.lower() if c.isalnum() else " " for c in text)
 
    return [
        tok
        for tok in cleaned.split()
        if len(tok) >= min_length and tok not in STOP_WORDS
    ]
 
 
def term_frequencies(tokens: list[str]) -> dict[str, int]:
    """Return a mapping of token > count. Used at index time."""
    freq: dict[str, int] = {}
    for tok in tokens:
        freq[tok] = freq.get(tok, 0) + 1
    return freq
 
