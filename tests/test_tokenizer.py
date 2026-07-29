from app.index.tokenizer import STOP_WORDS, term_frequencies, tokenize


def test_tokenize_lowercases():
    tokens = tokenize("Hello World", 2)
    assert all(tok == tok.lower() for tok in tokens)


def test_tokenize_strips_punctuation():
    tokens = tokenize("hello, world! foo-bar.", 2)
    for tok in tokens:
        assert all(c.isalnum() for c in tok)


def test_tokenize_removes_stop_words():
    tokens = tokenize("the quick brown fox", 2)
    assert not any(tok in STOP_WORDS for tok in tokens)


def test_tokenize_respects_min_length():
    tokens = tokenize("a go rust python", 4)
    assert all(len(tok) >= 4 for tok in tokens)


def test_tokenize_expected_tokens():
    tokens = tokenize("The quick brown fox jumps", 2)
    assert tokens == ["quick", "brown", "fox", "jumps"]


def test_tokenize_empty_string():
    assert tokenize("", 2) == []


def test_term_frequencies():
    freq = term_frequencies(["go", "rust", "go", "python", "go"])
    assert freq["go"] == 3
    assert freq["rust"] == 1
    assert freq.get("missing", 0) == 0