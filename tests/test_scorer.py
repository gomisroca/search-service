from app.index.scorer import idf, tf, tfidf


def test_tf_zero_when_no_tokens():
    assert tf(5, 0) == 0.0


def test_tf_proportional():
    assert abs(tf(3, 10) - 0.3) < 1e-9


def test_idf_zero_when_no_documents():
    assert idf(0, 0) == 0.0


def test_idf_high_when_term_is_rare():
    rare = idf(1000, 1)
    common = idf(1000, 500)
    assert rare > common


def test_idf_smoothed_never_negative():
    assert idf(100, 100) >= 0


def test_tfidf_combines_correctly():
    score = tfidf(3, 10, 1000, 1)
    expected = tf(3, 10) * idf(1000, 1)
    assert abs(score - expected) < 1e-9