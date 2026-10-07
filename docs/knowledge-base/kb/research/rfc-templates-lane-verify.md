---
as_of: "2026-10-01"
created: "2026-10-01"
grammar: 1
informs:
- RFC-003
is_draft: false
provenance: agent-drafted
run: rfc-templates
scope:
- docs/knowledge-base/**
status: final
summary: "Verification lane: four IETF and KEP/kepval items re-checked against primaries, correcting the I-D-stage structure premise and confirming metadata validation."
tags:
- lane
- verification
- rfc-format
- ietf
title: Lane verify — four recorded-but-unverified RFC/IETF and KEP items (rfc-templates)
type: research
updated: "2026-10-01"
---
# Research: lane-verify — four recorded-but-unverified items (RFC/IETF + KEP/kepval)

Lane: lane-verify (read-only). **Observed 2026-10-01** (anchor: latest `kubernetes/enhancements` master
commit `883e9d16`, 2026-10-01T18:41:55Z). `source_check` was **not** registered in this lane's toolset, so
every quote below is a direct single-pass fetch of the primary (no second-model corroboration).

## 1. IETF proposal-stage structure — **CORRECTED**

Premise ("nothing at I-D stage beyond boilerplate + expiry + the RFC 7322 post-publication skeleton") is
wrong: the IETF publishes an I-D-stage required-content list.

- Required **sections**, authors.ietf.org *Required Content*: "Ensure that the I-D contains each of the
  following sections: IPR-Related Notices / Abstract / Introduction / Security Considerations / IANA
  Considerations / References / Authors' Addresses". Required header/elements: "The date the document was
  generated. / A list of authors/editors and their affiliations. / The group that originated the I-D (if
  any). / The intended status of the document. / The set of RFCs this I-D intends to update or replace, if
  any."
- Provenance is explicitly RFC-derived: "their content requirements are based on what is required in an
  RFC"; "IETF stream I-Ds submitted to the IESG must follow all of the guidance." (See RFC 7841.)
- Plaintext structural rules (same page, Format): TOC "must not be numbered" and must sit "between the
  'Copyright Notice' and the introduction"; "the 'Status of This Memo' and 'Abstract' sections are not
  numbered"; Contributors/Acknowledgements "unnumbered and placed after the 'References' section".
- RFC 7322 §4 is a **post-publication** skeleton and disclaims hard structure: "A published RFC will
  largely contain the elements in the following list. Some of these sections are required, as noted.";
  "Within the body of the memo, the order shown above is strongly recommended... The section numbers above
  are for illustrative purposes; they are not intended to correspond to required numbering in an RFC."
- Partly machine-enforced: "If an Internet-Draft is prepared in XML, the tooling will ensure that required
  content is present"; but "The submission tool will double-check many, but not all, of the guidance."

**Verdict: CORRECTED.** A required section *set* exists at I-D stage (authors.ietf.org, inherited from RFC
requirements); RFC 7322 supplies no I-D-stage structure. No I-D document mandates an ordered/numbered
proposal skeleton beyond the plaintext TOC rules. *Design implication:* mandating a fixed section skeleton
for an agent-consumed RFC format is a legitimate strengthening — cite authors.ietf.org required-content as
the precedent, not RFC 7322.

## 2. kepval validation surface — **CORRECTED**

CI path traced: `hack/verify-kep-metadata.sh` → `go test -v ./test/metadata_test.go` → `Repo.Validate()`
→ `validateFile()` → `KEPHandler.Parse()` + `kepval.ValidatePRR()`. Current `pkg/kepval/` holds only
`approval.go`; `pkg/repo/validate.go` was last changed 2023-11-17.

Machine-checked:
1. **kep.yaml present for every README-based KEP** — "There is a proposal, we require metadata file for
   it." → `"metadata file missing for KEP: %s"` (`pkg/repo/validate.go`).
2. **Strict YAML, no unknown keys** — `yaml.UnmarshalStrict` (`api/proposal.go`); real failure text seen
   in-project: `line 26: field "foo" not found in type keps.Proposal`.
3. **Required keys (struct tags)** — `validate:"required"` on exactly `title`, `kep-number`, `authors`,
   `owning-sig`, `approvers`, `status`. `reviewers`, `stage`, `latest-milestone`, `milestone`,
   `feature-gates`, `metrics`, `disable-supported` are **not** required.
4. **PRR gating, conditional** — `isPRRRequired`: `required = kep.Status == api.ImplementableStatus ||
   kep.Status == api.ImplementedStatus`; then requires `latest-milestone` + `stage` ("missing the latest
   milestone field", "missing the stage field"), requires `prod-readiness/<owning-sig>/<kep-number>.yaml`,
   and requires an approver for the KEP's stage from the PRR approver list ("this contributor (%s) is not a
   PRR approver (%v)"). Gated on `latest-milestone >= v1.21`.
5. Enum/SIG/consistency checks **exist** in `api/proposal.go` `KEPHandler.Validate()` — `Status.IsValid()`
   (`provisional|implementable|implemented|deferred|rejected|withdrawn|replaced`), `Stage.IsValid()`
   (`alpha|beta|stable|deprecated|disabled|removed`), `validateGroups` (SIGs from kubernetes/community),
   `status:implemented implies stage:stable`. **Researcher inference (medium):** the CI path calls
   `KEPHandler.Parse()` (only `validateStruct`) and never `KEPHandler.Validate()`, so (5) may not run in
   the metadata presubmit. Not proven repo-wide.

Not machine-checked: README.md section presence; the PRR questionnaire *answers*; README↔kep.yaml
consistency. README *shape* is only advisory — `hack/verify-toc-vs-template.sh` diffs README headings
against `keps/NNNN-kep-template/README.md` (`mdtoc --max-depth 100`, ignoring removed `(Optional)`
headings) and ends `exit 0` with `# TODO(soltysh): for now this should not fail, but print problems`.

**Verdict: CORRECTED.** kepval = strict schema + 6 required keys + metadata-file presence + conditional PRR
approver gating; the prose body, questionnaire answers, and enum/SIG consistency are convention or advisory.
*Design implication:* machine checking only covers the typed control file — agents must not infer that
prose was validated.

## 3. RFC 7322 abstract/citation rules — **CONFIRMED (wording corrected)**

§4.3: "Note also that an Abstract is not a substitute for an Introduction; the RFC should be self-contained
as if there were no Abstract. Similarly, the Abstract should be complete in itself. It will appear in
isolation in publication announcements and in the online index of RFCs. Therefore, the Abstract must not
contain citations." §4.8: "The body of the memo and the Abstract must be self-contained and separable."
§4.8.6.4: "References to Internet-Drafts may only appear as informative references. Given that several
revisions of an I-D may be produced in a short time frame, references must include the posting date (month
and year), the full Internet-Draft file name (including the version number), and the phrase 'Work in
Progress'." Example: `Flanagan, H. and S. Ginoza, "RFC Style Guide", Work in Progress,
draft-flanagan-style-01, June 2013.`

Wording correction: "self-contained" in §4.3 describes **the RFC**, and the causal link is "complete in
itself / appears in isolation → therefore no citations" (not "self-contained because it appears in
isolation"). authors.ietf.org relaxes it for I-Ds: "An abstract should be complete in itself, so it should
not contain citations **unless they are completely defined within the abstract**."
*Design implication:* "no citations in abstract; expand abbreviations; 50–150 words" can be enforced at
authoring time — the IETF already states these as review checks.

## 4. I-D expiry (six months vs 185 days) — **CONFIRMED: acknowledged and reconciled**

- Legacy *Guidelines to Authors of Internet-Drafts* (still served as `1id-guidelines.txt`) mandates the
  verbatim boilerplate "valid for a maximum of six months..." (§5) **while in the same document** stating
  "The expiration date is 185 days following the I-D submission of the document. **Use of the phrase
  'expires in six months' or 'expires in 185 days' is not acceptable.**" and (§8) "An Internet-Draft will
  expire exactly 185 days from the date that it is posted on the IETF Web site".
- Current guidance (authors.ietf.org, After Submission): "An I-D expires 185 days after it was placed in
  the Repository, unless it is in a state that prevents it from expiring."
- Current `required-content.md` keeps the frozen "maximum of six months" boilerplate *and* requires item 3,
  "A statement specifying the expiry date of the Internet-Draft" (`This Internet-Draft will expire on DD
  MMMM YYYY.`), "automatically generated for authors who submit their I-D in RFCXML" — so the operative
  rule is the generated 185-day date; "six months" survives only as frozen text.
- An IETF-stream draft names the split outright: "The Content Guidelines for Internet Drafts [IDCG]
  requires that Internet-Drafts include an expiration statement. Tooling and IETF practice insist on
  Internet-Drafts including an expiry date 185 days after their posting."
  (draft-thomson-gendispatch-no-expiry; it and draft-levine-iduse propose deleting expiry notices.)

**Verdict: CONFIRMED** — acknowledged and operationally reconciled: 185 days is the rule, "six months" is
frozen boilerplate (no erratum; the fix path is policy/tooling). *Researcher inference:* 185 ≈ 6 × 30.4,
i.e. legacy rounding. *Design implication:* use an explicit computed date, never the "six months" phrasing.

## Contradictions
- **RFC 7322 vs itself:** §4.8.6.4 "References to Internet-Drafts may only appear as informative
  references", yet §4.8.6 "Normative references to Internet-Drafts will cause publication of the RFC to be
  suspended until the referenced draft is also ready for publication." Recorded, not resolved.
- **kepval code vs enforcement path** for the enum/SIG checks (item 2.5, inference).
- Items 3 and 4: none found.

## Missing evidence
- authors.ietf.org's "templates" page was not fetched; no canonical *I-D* section-*order* template located.
- No whole-repo grep of kubernetes/enhancements: "nothing else validates README contents" is bounded by the
  traced paths (`pkg/repo/*`, `pkg/kepval/*`, `api/proposal.go`, `hack/verify-*.sh`, `test/`, `cmd/`).
- No IETF errata/datatracker search for the expiry wording; "no erratum" is limited to this check.
- No `source_check` corroboration available (tool not registered in this lane).

## Sources
- Kept: RFC 7322 (rfc-editor.org/rfc/rfc7322.txt) — §3.5, §4 preamble, §4.3, §4.8, §4.8.6, §4.8.6.4.
- Kept: authors.ietf.org Required Content + `ietf/authors.ietf.org` `required-content.md`.
- Kept: "Guidelines to Authors of Internet-Drafts" (ietf.github.io/id-guidelines = authors.ietf.org).
- Kept: legacy `1id-guidelines.txt` — six-month boilerplate and 185-day rule in one document.
- Kept: kubernetes/enhancements `pkg/repo/{validate,repo}.go`, `pkg/kepval/approval.go`, `api/proposal.go`,
  `test/metadata_test.go`, `hack/verify-kep-metadata.sh`, `hack/verify-toc-vs-template.sh`, `docs/kepval.md`.
- Kept: draft-thomson-gendispatch-no-expiry(-01/-03).
- Rejected/deprioritized: kubernetes/enhancements PR/issue commentary (#1126, #1766, #2184, #4338, #2470) —
  history, not current behaviour; `chuckha/kepval` — superseded predecessor; community
  production-readiness.md — process narrative.

## Next steps
1. One full-repo grep for `KEPHandler.Validate` call sites to settle item 2.5.
2. Optional: check RFC 2026 §2.2 errata before quoting "six months" as anything but frozen boilerplate.
