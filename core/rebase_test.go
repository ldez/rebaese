package core

import "testing"

func Test_Should_create_an_SSH_URL_when_SSH_is_asked(t *testing.T) {
	cloneURL := "git://toto.com/foo/bar.git"

	repoURL := createRepositoryURL(cloneURL, true, "")

	if repoURL != cloneURL {
		t.Errorf("Expected %s but got %s", cloneURL, repoURL)
	}
}

func Test_Should_create_an_HTTPS_URL_when_SSH_is_not_asked(t *testing.T) {
	cloneURL := "git://toto.com/foo/bar.git"

	repoURL := createRepositoryURL(cloneURL, false, "")

	expectedURL := "https://toto.com/foo/bar.git"
	if repoURL != expectedURL {
		t.Errorf("Expected %s but got %s", expectedURL, repoURL)
	}
}

func Test_Should_create_an_HTTPS_URL_when_SSH_is_not_asked_and_provide_token(t *testing.T) {
	cloneURL := "git://toto.com/foo/bar.git"

	repoURL := createRepositoryURL(cloneURL, false, "secret")

	expectedURL := "https://secret@toto.com/foo/bar.git"
	if repoURL != expectedURL {
		t.Errorf("Expected %s but got %s", expectedURL, repoURL)
	}
}
