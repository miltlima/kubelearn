# Contributing to KubeLearn

Contributions are welcome. New questions are the easiest way to start: each one is a small, self-contained change.

## Adding a question

1. Write its checker in the package of its track under `pkg/resources/` (`easy`, `medium` and `hard` for Core, `cks`, `ckne`). Use `utils.Report` to collect the reasons the question is not solved, so learners get useful feedback.
2. Append it to `Questions` in `pkg/resources/registry.go` with its track. The position defines the question number, so add new questions at the end.
3. Add the objects that solve it to the fixtures in `pkg/resources/checks_test.go`. The tests then check that it passes when solved, fails on an empty cluster, and never panics on incomplete objects.
4. If it needs lab resources, add them to a file in `manifests/`.
5. Add the Portuguese and Spanish statements to `kubelearn-frontend/src/i18n.js`, keyed by the question number. Without them the question is shown in English.
6. Add an example answer to `manifests_examples/` if the question is not trivial.

A checker reads the objects in the cluster. It cannot send traffic or inspect nodes, so pick tasks whose result is visible in the Kubernetes API.

## Running the checks

```sh
go vet ./...
go test ./...
npm --prefix kubelearn-frontend test -- --watchAll=false
```

The same checks run on every pull request.

## Other changes

- **Bugs and ideas:** open an issue describing what you expected and what happened.
- **Larger changes:** open an issue first, so the approach can be agreed before you write the code.
- **Translations:** corrections to the Portuguese and Spanish texts, and new languages, go in `kubelearn-frontend/src/i18n.js`.

By contributing you agree that your work is licensed under the Apache License 2.0.
