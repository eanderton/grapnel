package toml

import "testing"

func testFlow(t *testing.T, input string, expectedFlow []token) {
	_, ch := lex(input)
	for _, expected := range expectedFlow {
		token := <-ch
		if token != expected {
			t.Log("While testing: ", input)
			t.Log("compared", token, "to", expected)
			t.Log(token.val, "<->", expected.val)
			t.Log(token.typ, "<->", expected.typ)
			t.Log(token.Line, "<->", expected.Line)
			t.Log(token.Col, "<->", expected.Col)
			t.FailNow()
		}
	}

	tok, ok := <-ch
	if ok {
		t.Log("channel is not closed!")
		t.Log(len(ch)+1, "tokens remaining:")

		t.Log("token ->", tok)
		for token := range ch {
			t.Log("token ->", token)
		}
		t.FailNow()
	}
}

func TestValidKeyGroup(t *testing.T) {
	testFlow(t, "[hello world]", []token{
		token{Position{0, 0}, tokenLeftBracket, "["},
		token{Position{0, 1}, tokenKeyGroup, "hello world"},
		token{Position{0, 12}, tokenRightBracket, "]"},
		token{Position{0, 13}, tokenEOF, ""},
	})
}

func TestUnclosedKeyGroup(t *testing.T) {
	testFlow(t, "[hello world", []token{
		token{Position{0, 0}, tokenLeftBracket, "["},
		token{Position{0, 1}, tokenError, "unclosed key group"},
	})
}

func TestComment(t *testing.T) {
	testFlow(t, "# blahblah", []token{
		token{Position{0, 10}, tokenEOF, ""},
	})
}

func TestKeyGroupComment(t *testing.T) {
	testFlow(t, "[hello world] # blahblah", []token{
		token{Position{0, 0}, tokenLeftBracket, "["},
		token{Position{0, 1}, tokenKeyGroup, "hello world"},
		token{Position{0, 12}, tokenRightBracket, "]"},
		token{Position{0, 24}, tokenEOF, ""},
	})
}

func TestMultipleKeyGroupsComment(t *testing.T) {
	testFlow(t, "[hello world] # blahblah\n[test]", []token{
		token{Position{0, 0}, tokenLeftBracket, "["},
		token{Position{0, 1}, tokenKeyGroup, "hello world"},
		token{Position{0, 12}, tokenRightBracket, "]"},
		token{Position{1, 0}, tokenLeftBracket, "["},
		token{Position{1, 1}, tokenKeyGroup, "test"},
		token{Position{1, 5}, tokenRightBracket, "]"},
		token{Position{1, 6}, tokenEOF, ""},
	})
}

func TestBasicKey(t *testing.T) {
	testFlow(t, "hello", []token{
		token{Position{0, 0}, tokenKey, "hello"},
		token{Position{0, 5}, tokenEOF, ""},
	})
}

func TestBasicKeyWithUnderscore(t *testing.T) {
	testFlow(t, "hello_hello", []token{
		token{Position{0, 0}, tokenKey, "hello_hello"},
		token{Position{0, 11}, tokenEOF, ""},
	})
}

func TestBasicKeyWithDash(t *testing.T) {
	testFlow(t, "hello-world", []token{
		token{Position{0, 0}, tokenKey, "hello-world"},
		token{Position{0, 11}, tokenEOF, ""},
	})
}

func TestBasicKeyWithUppercaseMix(t *testing.T) {
	testFlow(t, "helloHELLOHello", []token{
		token{Position{0, 0}, tokenKey, "helloHELLOHello"},
		token{Position{0, 15}, tokenEOF, ""},
	})
}

func TestBasicKeyWithInternationalCharacters(t *testing.T) {
	testFlow(t, "héllÖ", []token{
		token{Position{0, 0}, tokenKey, "héllÖ"},
		token{Position{0, 5}, tokenEOF, ""},
	})
}

func TestBasicKeyAndEqual(t *testing.T) {
	testFlow(t, "hello =", []token{
		token{Position{0, 0}, tokenKey, "hello"},
		token{Position{0, 6}, tokenEqual, "="},
		token{Position{0, 7}, tokenEOF, ""},
	})
}

func TestKeyWithSharpAndEqual(t *testing.T) {
	testFlow(t, "key#name = 5", []token{
		token{Position{0, 0}, tokenKey, "key#name"},
		token{Position{0, 9}, tokenEqual, "="},
		token{Position{0, 11}, tokenInteger, "5"},
		token{Position{0, 12}, tokenEOF, ""},
	})
}

func TestKeyWithSymbolsAndEqual(t *testing.T) {
	testFlow(t, "~!@#$^&*()_+-`1234567890[]\\|/?><.,;:' = 5", []token{
		token{Position{0, 0}, tokenKey, "~!@#$^&*()_+-`1234567890[]\\|/?><.,;:'"},
		token{Position{0, 38}, tokenEqual, "="},
		token{Position{0, 40}, tokenInteger, "5"},
		token{Position{0, 41}, tokenEOF, ""},
	})
}

func TestKeyEqualStringEscape(t *testing.T) {
	testFlow(t, `foo = "hello\""`, []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 7}, tokenString, "hello\""},
		token{Position{0, 15}, tokenEOF, ""},
	})
}

func TestKeyEqualStringUnfinished(t *testing.T) {
	testFlow(t, `foo = "bar`, []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 7}, tokenError, "unclosed string"},
	})
}

func TestKeyEqualString(t *testing.T) {
	testFlow(t, `foo = "bar"`, []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 7}, tokenString, "bar"},
		token{Position{0, 11}, tokenEOF, ""},
	})
}

func TestKeyEqualTrue(t *testing.T) {
	testFlow(t, "foo = true", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenTrue, "true"},
		token{Position{0, 10}, tokenEOF, ""},
	})
}

func TestKeyEqualFalse(t *testing.T) {
	testFlow(t, "foo = false", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenFalse, "false"},
		token{Position{0, 11}, tokenEOF, ""},
	})
}

func TestArrayNestedString(t *testing.T) {
	testFlow(t, `a = [ ["hello", "world"] ]`, []token{
		token{Position{0, 0}, tokenKey, "a"},
		token{Position{0, 2}, tokenEqual, "="},
		token{Position{0, 4}, tokenLeftBracket, "["},
		token{Position{0, 6}, tokenLeftBracket, "["},
		token{Position{0, 8}, tokenString, "hello"},
		token{Position{0, 14}, tokenComma, ","},
		token{Position{0, 17}, tokenString, "world"},
		token{Position{0, 23}, tokenRightBracket, "]"},
		token{Position{0, 25}, tokenRightBracket, "]"},
		token{Position{0, 26}, tokenEOF, ""},
	})
}

func TestArrayNestedInts(t *testing.T) {
	testFlow(t, "a = [ [42, 21], [10] ]", []token{
		token{Position{0, 0}, tokenKey, "a"},
		token{Position{0, 2}, tokenEqual, "="},
		token{Position{0, 4}, tokenLeftBracket, "["},
		token{Position{0, 6}, tokenLeftBracket, "["},
		token{Position{0, 7}, tokenInteger, "42"},
		token{Position{0, 9}, tokenComma, ","},
		token{Position{0, 11}, tokenInteger, "21"},
		token{Position{0, 13}, tokenRightBracket, "]"},
		token{Position{0, 14}, tokenComma, ","},
		token{Position{0, 16}, tokenLeftBracket, "["},
		token{Position{0, 17}, tokenInteger, "10"},
		token{Position{0, 19}, tokenRightBracket, "]"},
		token{Position{0, 21}, tokenRightBracket, "]"},
		token{Position{0, 22}, tokenEOF, ""},
	})
}

func TestArrayInts(t *testing.T) {
	testFlow(t, "a = [ 42, 21, 10, ]", []token{
		token{Position{0, 0}, tokenKey, "a"},
		token{Position{0, 2}, tokenEqual, "="},
		token{Position{0, 4}, tokenLeftBracket, "["},
		token{Position{0, 6}, tokenInteger, "42"},
		token{Position{0, 8}, tokenComma, ","},
		token{Position{0, 10}, tokenInteger, "21"},
		token{Position{0, 12}, tokenComma, ","},
		token{Position{0, 14}, tokenInteger, "10"},
		token{Position{0, 16}, tokenComma, ","},
		token{Position{0, 18}, tokenRightBracket, "]"},
		token{Position{0, 19}, tokenEOF, ""},
	})
}

func TestMultilineArrayComments(t *testing.T) {
	testFlow(t, "a = [1, # wow\n2, # such items\n3, # so array\n]", []token{
		token{Position{0, 0}, tokenKey, "a"},
		token{Position{0, 2}, tokenEqual, "="},
		token{Position{0, 4}, tokenLeftBracket, "["},
		token{Position{0, 5}, tokenInteger, "1"},
		token{Position{0, 6}, tokenComma, ","},
		token{Position{1, 0}, tokenInteger, "2"},
		token{Position{1, 1}, tokenComma, ","},
		token{Position{2, 0}, tokenInteger, "3"},
		token{Position{2, 1}, tokenComma, ","},
		token{Position{3, 0}, tokenRightBracket, "]"},
		token{Position{3, 1}, tokenEOF, ""},
	})
}

func TestKeyEqualArrayBools(t *testing.T) {
	testFlow(t, "foo = [true, false, true]", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenLeftBracket, "["},
		token{Position{0, 7}, tokenTrue, "true"},
		token{Position{0, 11}, tokenComma, ","},
		token{Position{0, 13}, tokenFalse, "false"},
		token{Position{0, 18}, tokenComma, ","},
		token{Position{0, 20}, tokenTrue, "true"},
		token{Position{0, 24}, tokenRightBracket, "]"},
		token{Position{0, 25}, tokenEOF, ""},
	})
}

func TestKeyEqualArrayBoolsWithComments(t *testing.T) {
	testFlow(t, "foo = [true, false, true] # YEAH", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenLeftBracket, "["},
		token{Position{0, 7}, tokenTrue, "true"},
		token{Position{0, 11}, tokenComma, ","},
		token{Position{0, 13}, tokenFalse, "false"},
		token{Position{0, 18}, tokenComma, ","},
		token{Position{0, 20}, tokenTrue, "true"},
		token{Position{0, 24}, tokenRightBracket, "]"},
		token{Position{0, 32}, tokenEOF, ""},
	})
}

func TestDateRegexp(t *testing.T) {
	if dateRegexp.FindString("1979-05-27T07:32:00Z") == "" {
		t.Fail()
	}
}

func TestKeyEqualDate(t *testing.T) {
	testFlow(t, "foo = 1979-05-27T07:32:00Z", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenDate, "1979-05-27T07:32:00Z"},
		token{Position{0, 26}, tokenEOF, ""},
	})
}

func TestFloatEndingWithDot(t *testing.T) {
	testFlow(t, "foo = 42.", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenError, "float cannot end with a dot"},
	})
}

func TestFloatWithTwoDots(t *testing.T) {
	testFlow(t, "foo = 4.2.", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenError, "cannot have two dots in one float"},
	})
}

func TestDoubleEqualKey(t *testing.T) {
	testFlow(t, "foo= = 2", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 3}, tokenEqual, "="},
		token{Position{0, 4}, tokenError, "cannot have multiple equals for the same key"},
	})
}

func TestInvalidEsquapeSequence(t *testing.T) {
	testFlow(t, `foo = "\x"`, []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 7}, tokenError, "invalid escape sequence: \\x"},
	})
}

func TestNestedArrays(t *testing.T) {
	testFlow(t, "foo = [[[]]]", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenLeftBracket, "["},
		token{Position{0, 7}, tokenLeftBracket, "["},
		token{Position{0, 8}, tokenLeftBracket, "["},
		token{Position{0, 9}, tokenRightBracket, "]"},
		token{Position{0, 10}, tokenRightBracket, "]"},
		token{Position{0, 11}, tokenRightBracket, "]"},
		token{Position{0, 12}, tokenEOF, ""},
	})
}

func TestKeyEqualNumber(t *testing.T) {
	testFlow(t, "foo = 42", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenInteger, "42"},
		token{Position{0, 8}, tokenEOF, ""},
	})

	testFlow(t, "foo = +42", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenInteger, "+42"},
		token{Position{0, 9}, tokenEOF, ""},
	})

	testFlow(t, "foo = -42", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenInteger, "-42"},
		token{Position{0, 9}, tokenEOF, ""},
	})

	testFlow(t, "foo = 4.2", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenFloat, "4.2"},
		token{Position{0, 9}, tokenEOF, ""},
	})

	testFlow(t, "foo = +4.2", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenFloat, "+4.2"},
		token{Position{0, 10}, tokenEOF, ""},
	})

	testFlow(t, "foo = -4.2", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenFloat, "-4.2"},
		token{Position{0, 10}, tokenEOF, ""},
	})
}

func TestMultiline(t *testing.T) {
	testFlow(t, "foo = 42\nbar=21", []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 6}, tokenInteger, "42"},
		token{Position{1, 0}, tokenKey, "bar"},
		token{Position{1, 3}, tokenEqual, "="},
		token{Position{1, 4}, tokenInteger, "21"},
		token{Position{1, 6}, tokenEOF, ""},
	})
}

func TestKeyEqualStringUnicodeEscape(t *testing.T) {
	testFlow(t, `foo = "hello \u2665"`, []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 7}, tokenString, "hello ♥"},
		token{Position{0, 20}, tokenEOF, ""},
	})
}

func TestUnicodeString(t *testing.T) {
	testFlow(t, `foo = "hello ♥ world"`, []token{
		token{Position{0, 0}, tokenKey, "foo"},
		token{Position{0, 4}, tokenEqual, "="},
		token{Position{0, 7}, tokenString, "hello ♥ world"},
		token{Position{0, 21}, tokenEOF, ""},
	})
}

func TestKeyGroupArray(t *testing.T) {
	testFlow(t, "[[foo]]", []token{
		token{Position{0, 0}, tokenDoubleLeftBracket, "[["},
		token{Position{0, 2}, tokenKeyGroupArray, "foo"},
		token{Position{0, 5}, tokenDoubleRightBracket, "]]"},
		token{Position{0, 7}, tokenEOF, ""},
	})
}
