# Source-Available Licenses for Combined Hardware-Software Projects

**Research Date:** 2026-07-15  
**Mode:** Standard  
**Question:** Which source-available or fair-code license allows a combined hardware + software project (similar to 3D printers) to remain publicly accessible for community use while preventing third parties from copying and selling the product or software without the author's authorization?

---

## Executive Summary

- **Key Finding 1:** The **Open Community License (OCL) v1.1**, created by Prusa Research for its CORE One 3D printer line, is the only modern public license explicitly drafted to cover **both hardware and software** under a single instrument while restricting commercial sale of the product or derivatives without a separate business license [1][2].
- **Key Finding 2:** No non-commercial or sale-restricted license qualifies as "open source" under the OSI Open Source Definition or the OSHWA Open Source Hardware Definition, because both require the right to **make, sell, and distribute** derivatives without field-of-use restrictions [4][14].
- **Key Finding 3:** For software-only portions, **PolyForm Noncommercial 1.0.0** and **Commons Clause** are mature source-available alternatives, but neither covers physical hardware designs [5][7].
- **Key Finding 4:** Traditional open hardware licenses such as **CERN OHL** (all variants) and **TAPR OHL** are designed for hardware documentation and either permit commercial sale outright (CERN-OHL-P) or require reciprocity but still allow sale (CERN-OHL-S/W, TAPR OHL); they therefore do not block third-party cloning [8][9].
- **Key Finding 5:** Relicensing an existing MIT project requires **consent from all copyright holders** unless the new terms are compatible with MIT; previous releases remain under MIT in perpetuity [10][14].

**Primary Recommendation:** Adopt **OCL v1.1** for the combined hardware-software RadKeys project, add the **Micro Business plugin** to allow small companies free internal use, publish a separate **commercial licensing contact procedure**, and clearly avoid calling the project "open source" in the strict OSI/OSHWA sense.

**Confidence Level:** High for factual descriptions of each license; Medium for enforceability predictions, because OCL is new and untested in courts.

---

## Introduction

### Research Question

The RadKeys project currently uses the MIT License. The author wants to keep the repository publicly accessible to attract a community, but prevent third parties from manufacturing, selling, or distributing the product (hardware + firmware + host software) without explicit authorization. The desired model is to negotiate commercial agreements with companies, influencers, and resellers in exchange for royalties or commissions. The research therefore asks: **which public license is used by hardware-software projects (notably 3D printing) to achieve exactly this balance?**

### Scope & Methodology

This research investigates publicly available licenses that (a) cover combined hardware and software, (b) allow non-commercial community access, modification, and repair, and (c) restrict third-party commercial sale of the product or its derivatives without a separate agreement. The investigation focused on the 3D-printing ecosystem because it is the most visible hardware-software domain that recently produced a license matching these requirements.

Methods used:
- Web search across general, news, and academic indexes for "3D printer non-commercial hardware-software license", "Open Community License Prusa", "source available hardware license", and related terms.
- Direct reading of primary license texts and authoritative repositories (OCL, PolyForm, Commons Clause, CERN OHL, TAPR OHL, OSHWA definition, MIT).
- Legal commentary and practitioner analysis from licensing attorneys and community critics.
- Academic survey paper on open hardware licensing.

**Sources consulted:** 16 primary and secondary sources, spanning license stewards, project maintainers, legal analysts, and standard-setting organizations. Date range: 2007 (TAPR NCL) to June 2026 (Prusa OCL v1.1 announcement).

### Key Assumptions

- **Assumption 1:** The goal is to restrict *sale of the product/software as a product*, not to block all commercial use (e.g., internal business use as a tool may remain allowed).
- **Assumption 2:** The project is willing to abandon the "open source" label in the strict OSI/OSHWA sense, since any non-commercial or sale restriction violates those definitions.
- **Assumption 3:** The author either owns all copyrights or can secure contributor consent before relicensing.
- **Assumption 4:** The hardware is a tangible device with firmware running on it and a host configurator application, so a single license covering both hardware documentation and software is preferable to dual licensing.

---

## Main Analysis

### Finding 1: Open Community License v1.1 Is the Closest Match

Prusa Research introduced the Open Community License (OCL) in 2025 alongside the CORE One 3D printer CAD files, then released **OCL v1.1** on 24 June 2026 with a modular plugin system [2]. The license was explicitly created because traditional open source licenses "left gaps" when applied to physical hardware: they allowed competitors to copy and sell machines without sharing revenue or contributing back [16].

The OCL repository states that the license "pertains to intellectual property applied in licensor's product (licensed item - hardware and/or software) and/or its components which is distributed as is under OCL, including copyright, registered design or design patent, and (utility) patent" [1]. This is unusual: most public licenses cover either software (copyright + patent) or hardware documentation, but OCL explicitly combines both in one instrument [3].

For non-commercial end users, OCL grants broad rights: "use, copy, modify, hack the product and/or its components as you wish" [1]. Derivatives distributed publicly must remain under OCL, preventing license fragmentation [2]. For business users, OCL allows internal business use and internal production, but prohibits copying or replicating the product or its derivatives for any commercial purpose without a separate business or repair license [1].

The v1.1 plugin system adds four optional modules [2]:
- **GAtt v1** — general attribution requirements for derivatives.
- **SWAtt v1** — software-specific attribution in UI and source code.
- **Micro v1** — permits free commercial internal use for entities with annual gross revenue under €1 million, including affiliates.
- **RnD v1** — restricts business use to research and development; manufacturing requires a separate agreement.

For RadKeys, the **Micro v1** plugin is particularly relevant: it would let small businesses, startups, and individual makers use the device internally without friction, while triggering a licensing conversation only once a reseller or manufacturer reaches meaningful scale [2].

Practical examples published by Prusa clarify that using an OCL-licensed machine as a tool to manufacture one's own products is allowed; only copying or replicating the OCL-licensed design for sale is restricted [2]. This aligns with the RadKeys use case: radiologists can use the device, hobbyists can build one, but a company cannot clone and sell the keypad or configurator without permission.

**Sources:** [1], [2], [3], [16]

---

### Finding 2: OCL Is Not "Open Source" Under OSI/OSHWA Definitions

A critical limitation of OCL is that it cannot be called "open source" in the sense recognized by the Open Source Initiative or the Open Source Hardware Association. The OSHWA definition requires a license to "allow for the manufacture, sale, distribution, and use of products created from the design files, the design files themselves, and derivatives thereof" [4]. OCL explicitly restricts sale of products and derivatives for commercial purposes without a separate license, which directly contradicts this requirement.

OSHWA also states that licenses carrying restrictions against commercial use or derivative works are "strictly incompatible with open source" [4]. The OCL repository itself acknowledges that OCL is "designed to allow open access to the relevant community" and promote "fair competition," but it does not claim OSI or OSHWA approval [1].

Legal analyst Kyle Mitchell reviewed OCL v1 and concluded that the combination of commercial/non-commercial segmentation plus share-alike terms makes it "particularly complex, functionally speaking," and noted the lack of definitions for "commercial" or "non-commercial" as a drafting weakness [3]. Mitchell's practical takeaway: "I probably wouldn't hesitate to rely on this license as a hobbyist doing things no one would confuse for business. Everyone else should have a look and think twice" [3].

4Additive summarized the community view: OCL "allows free use and modification for non-commercial purposes, but limits companies to internal production only, prohibiting the sale of derivatives. The community considers it incompatible with open source hardware standards" [15].

This means RadKeys must be careful with marketing language. Terms such as **source-available**, **community license**, **fair-code hardware**, or **open design** are more accurate than "open source" if OCL is adopted.

**Sources:** [3], [4], [15]

---

### Finding 3: Software-Only Source-Available Licenses Do Not Cover Hardware

If RadKeys were to split licensing by artifact type, several well-established source-available licenses exist for the software portions, but none are designed for hardware designs.

**PolyForm Noncommercial 1.0.0** is a plain-language software license that permits any noncommercial purpose, personal use, and use by charitable, educational, public research, public safety, health, environmental, and government institutions regardless of funding source [5]. It explicitly grants copyright and patent licenses, but only for noncommercial purposes. PolyForm is maintained by experienced licensing lawyers and is listed on SPDX [5].

**Commons Clause** is a short condition that can be appended to an OSI-approved license such as Apache 2.0. It removes the right to "Sell" the software, defined as providing a product or service whose value derives "entirely or substantially" from the software's functionality, for a fee [7]. The Commons Clause FAQ explicitly states: "Is this 'Open Source'? No" [7]. It also notes that Commons Clause is narrower than Creative Commons NonCommercial because it restricts only one kind of commercial use rather than all commercial use [7].

**Fair-code** is not a license but a model describing software that is free to use and modify, has source available, and is commercially restricted by its authors [6]. Fair-code-compatible licenses include the Business Source License, Commons Clause, Confluent Community License, Elastic License 2.0, Server Side Public License, and Sustainable Use License [6]. These are used by companies such as MongoDB, Elastic, HashiCorp, n8n, and Sentry [6].

However, none of these software licenses clearly cover hardware documentation, CAD files, PCB layouts, or physical products. Applying them to hardware would create legal uncertainty. Therefore, they are best viewed as complements (for a software-only submodule) rather than replacements for a hardware-covering license.

**Sources:** [5], [6], [7]

---

### Finding 4: Traditional Open Hardware Licenses Permit Sale and Cloning

The most widely used hardware-specific licenses — CERN OHL v2 and TAPR OHL — were designed to maximize sharing and reuse, not to block commercial cloning.

CERN OHL v2 comes in three variants [8]:
- **CERN-OHL-P (permissive):** allows proprietary derivatives, does not require source disclosure, and permits commercial sale.
- **CERN-OHL-W (weakly reciprocal):** requires sharing modifications to the licensed design, but allows larger works to remain proprietary.
- **CERN-OHL-S (strongly reciprocal):** requires sharing derivatives under the same license, similar to GPL for hardware.

All three explicitly allow commercial use. The SPDX summary for CERN-OHL-S lists "Commercial use, Distribution, Modification, Patent use, Private use" as permissions [8]. None of them restrict sale of products based on the design.

TAPR OHL v1.0 is a reciprocal hardware license. Its preamble states: "You may use products for any legal purpose without limitation" and "You may distribute products you make to third parties" [9]. TAPR also previously offered a **Noncommercial Hardware License (NCL)**, but it has been **deprecated** [9]. A Crabgrass group note records that TAPR deprecated the NCL after issues were communicated by Bruce Perens, making TAPR "100% Open Hardware now" [9]. Even when active, the TAPR NCL was designed for hardware documentation and did not clearly cover firmware or software loaded into programmable devices.

The Turing Way summarizes the landscape: CERN OHL v2 provides strongly reciprocal, weakly reciprocal, and permissive alternatives; TAPR OHL is also reciprocal; and Solderpad is permissive [13]. None of these provide a non-commercial or sale-restriction option.

For RadKeys, choosing CERN-OHL-S would keep derivatives open, but would explicitly allow competitors to manufacture and sell clones as long as they shared their own design files. That is the opposite of the desired protection.

**Sources:** [8], [9], [13]

---

### Finding 5: Relicensing from MIT Requires Contributor Consent

The current RadKeys repository is under MIT. Changing to OCL or any source-available license is legally possible only if the copyright holder(s) agree. The MIT License grants broad rights — including the right to "sell copies of the Software" — and those rights are perpetual for any version already released [14].

Open Source Stack Exchange consensus is clear: a sole author can relicense unilaterally; if there are multiple contributors, all must agree, or their contributions must be removed, or the new license must be compatible with the old one [10]. One answer notes that the Eclipse Foundation took more than a year to secure approvals when changing from CPL to EPL [10]. Another points out that even if one contributor refuses, their code cannot be relicensed without removal [10].

This has practical consequences for RadKeys:
- If the author is the only contributor, the switch is straightforward for future releases.
- If any third-party code exists under MIT, that code can remain under MIT in old releases; only new releases can be under OCL.
- Any MIT-licensed dependencies used by RadKeys are unaffected, because OCL explicitly states that components provided under incompatible licenses remain unaffected [1].

The author should document the license change prominently and retain evidence of sole authorship or contributor consent.

**Sources:** [10], [14]

---

### Finding 6: Commercial Licensing Models and Practical Implementation

Restricting sale is only half of the strategy; the other half is creating a clear path for authorized commercial use. Several projects illustrate the model.

**Open Source Ecology (OSE)** publishes a template licensing agreement under which OSE grants a non-exclusive global license to use its technology and brand, subject to branding guidelines [web search result, not registered].

**SMUPI** separates licensing into two streams: component manufacturers and module builders, with geographic exclusivity and quality standards [web search result, not registered].

**TetherIA's Aero Hand Open** uses a tripartite approach: software/firmware under Apache 2.0, hardware files under CC BY-NC-SA 4.0, and physical units governed by Terms of Sale plus a "Commercial Integration Permission" [web search result, not registered].

For RadKeys, a practical commercial licensing page should include:
1. **What is restricted:** manufacturing, assembly, distribution, or sale of the device, firmware, or configurator; offering them as a paid service; or creating derivatives for sale.
2. **What is allowed without negotiation:** personal use, internal business use, right-to-repair, modifications kept private, and derivatives shared under OCL without sale.
3. **How to request a license:** a contact email or form, required information (company, territory, expected volume, derivative or exact clone), and typical terms (royalty percentage, minimum guarantee, attribution, quality standards).
4. **Micro-business carve-out:** using OCL's Micro v1 plugin, free internal use below €1M annual gross revenue.

The arXiv survey on open hardware licensing notes that hybrid approaches with dual licenses — open for non-commercial sharing, proprietary for commercial uses — are common strategies for retaining business value [11]. OCL's plugin system effectively formalizes this hybrid approach in a single license family.

**Sources:** [2], [11]

---

## Synthesis & Insights

### Patterns Identified

**Pattern 1: The "source-available hardware" gap.** The 3D-printing industry has produced a new license category because neither software source-available licenses nor traditional open hardware licenses solved the cloning problem. Software licenses do not cover physical designs; hardware licenses permit sale. OCL is an attempt to bridge this gap.

**Pattern 2: Community acceptance trades off against commercial protection.** Every license that restricts commercial sale is explicitly excluded from the OSI/OSHWA "open source" definition. Projects adopting such licenses must choose their marketing vocabulary carefully.

**Pattern 3: Plugin modularity is emerging as a best practice.** OCL v1.1's plugin system reflects a broader trend: one core license with optional conditions (attribution, revenue thresholds, R&D-only) lets creators tailor terms without drafting custom licenses from scratch.

### Novel Insights

**Insight 1: OCL is less about "non-commercial" and more about "no cloning."** Unlike CC BY-NC-SA, which broadly restricts commercial use, OCL allows businesses to *use* the product internally and even manufacture their own parts for internal production. The restriction is on *replicating and selling the design or product*. This is closer to a no-cloning rule than a pure non-commercial rule, making it more compatible with real business adoption.

**Insight 2: The Micro v1 plugin solves the "small business" objection.** A common criticism of non-commercial licenses is that they harm tiny businesses and side hustles. By setting a €1M revenue threshold, OCL Micro lets small actors operate freely while reserving large-scale commercialization for negotiation.

**Insight 3: For RadKeys, the hardware-software unity matters more than for most projects.** Because the RP2040-Zero firmware and the Go/Fyne configurator are deeply coupled to the physical 36-button deck, a single license covering hardware, firmware, and software reduces legal fragmentation and user confusion.

### Implications

**For RadKeys:** Adopting OCL v1.1 + Micro v1 would:
- Keep the repository public and attract community contributors and testers.
- Allow individuals and small businesses to build and use the device freely.
- Block unlicensed manufacturers, resellers, and influencers from selling clones.
- Create a negotiation framework for authorized commercial partnerships.
- Require abandoning the strict "open source" label.

**Broader Implications:** If OCL gains adoption beyond Prusa, it could become a de facto standard for "open design, restricted sale" hardware-software products. However, its enforceability in court remains unproven, and its drafting ambiguities (noted by Mitchell) may create litigation risk.

**Second-Order Effects:** Competitors may still reverse-engineer the device and implement a clean-room alternative under a different license. OCL protects the *design files and software* but cannot protect the underlying idea, functionality, or unpatented hardware concepts. Trademark protection for the RadKeys name and logo remains important regardless of license choice.

---

## Limitations & Caveats

### Counterevidence Register

**Contradictory Finding 1: OCL drafting is immature.** Kyle Mitchell identified multiple ambiguities in OCL v1: undefined "commercial" and "non-commercial," awkward phrasing of "internal production use," and unclear status of the anti-data-mining paragraph [3]. These issues could weaken enforcement in court.
- Impact on conclusions: Moderate. OCL is still the best available option for the stated goals, but the author should expect to iterate and possibly seek legal review.

**Contradictory Finding 2: Some community voices reject OCL as neither open nor community-friendly.** Adafruit's legal analysis (cited in search results, 403-blocked from direct fetch) argues that calling a non-commercial restriction "Open Community License" is misleading. 4Additive reports that the community considers OCL incompatible with open hardware standards [15].
- Impact on conclusions: Moderate. Marketing and communication strategy must manage this reputational risk.

**Contradictory Finding 3: TAPR NCL was deprecated.** The existence of a prior non-commercial hardware license that was abandoned suggests the model has historical weaknesses, including possible incompatibility with community norms and enforcement difficulties [9].
- Impact on conclusions: Low. OCL differs from TAPR NCL by covering software and offering a plugin system; it is a more modern attempt.

### Known Gaps

- **Gap 1: No court decisions interpreting OCL.** Because OCL is new, there is no case law on enforceability, jurisdiction, or remedies. Predictions about how courts will interpret it are speculative.
- **Gap 2: No direct comparison of OCL to Brazilian law.** The author is in Brazil; OCL's enforceability under Brazilian copyright, industrial design, and contract law has not been analyzed in this research.
- **Gap 3: Patent and trademark strategy.** A license alone does not prevent competitors from designing around the product. The research did not assess patentability of the RadKeys hardware or the strength of trademark protection.

### Assumptions Revisited

- **Assumption 1 (restrict sale, not all use):** Supported. OCL's structure aligns with this assumption.
- **Assumption 2 (willing to abandon "open source" label):** Required. All evidence confirms that any sale restriction disqualifies the project from OSI/OSHWA definitions.
- **Assumption 3 (sole authorship or consent):** Must be verified by the author before proceeding.
- **Assumption 4 (single license preferred):** Supported. OCL is the only identified single-license solution.

### Areas of Uncertainty

- How Brazilian courts would treat a non-commercial restriction in a public license.
- Whether OCL's anti-data-mining clause is enforceable against AI training under local fair-use or text-and-data-mining exceptions.
- The exact boundary between "internal business use" and "commercial purpose" for a medical device used in a hospital or clinic.

---

## Recommendations

### Immediate Actions

1. **Verify authorship and obtain any necessary consent.**
   - What: Audit the Git history to confirm whether RadKeys has any third-party contributions under MIT.
   - Why: Relicensing requires consent from all copyright holders [10].
   - How: Run `git shortlog -sne` and review `git log --stat`. If contributors exist, email them for written consent or remove their code.
   - Timeline: Before changing any license file.

2. **Adopt OCL v1.1 + Micro v1 plugin.**
   - What: Replace `LICENSE` with the OCL v1.1 text and add `OCL v1.1 + Micro v1` to project metadata.
   - Why: This is the only identified license that covers hardware + software and restricts commercial sale while allowing small-business internal use [1][2].
   - How: Copy the authoritative text from the OCL repository [1]; add the Micro v1 add-on text from the `addons/` directory.
   - Timeline: After authorship verification.

3. **Create a `COMMERCIAL-LICENSE.md` file.**
   - What: Document what commercial activities require authorization and how to request it.
   - Why: A restriction without a clear path to authorization is commercially useless.
   - How: Include contact email, required information, standard terms outline, and response-time expectation.
   - Timeline: Same commit as license change.

4. **Update `README.md` and marketing language.**
   - What: Replace "open source" with "source-available," "community license," or "open design" where accurate.
   - Why: Prevents reputational backlash and avoids misrepresentation under OSI/OSHWA definitions [4][15].
   - How: Add a clear paragraph explaining: public for community, free for personal/small-business internal use, commercial sale requires license.
   - Timeline: Same commit.

### Next Steps (1–3 Months)

1. **Consult a Brazilian intellectual-property attorney.**
   - Review OCL enforceability under Brazilian law, especially for hardware designs and software.
   - Draft or adapt a commercial licensing agreement template in Portuguese and English.

2. **Register trademarks for "RadKeys" and logos.**
   - A license restricts copying of files; a trademark protects the brand against counterfeiters using the name.

3. **Publish a contribution policy and CLA.**
   - For future contributors, require a Contributor License Agreement that grants the project owner the right to license contributions under OCL and future versions, and to offer commercial licenses.

### Further Research Needs

1. **Brazilian IP law analysis.**
   - Specific enforceability of OCL's non-commercial and anti-cloning provisions in Brazil.
   - Whether Brazilian industrial design registration is advisable for the keypad enclosure.

2. **Comparison with custom proprietary/commercial license.**
   - Whether a fully custom license drafted by a Brazilian attorney would provide stronger protection than OCL, despite losing the "community license" signaling.

3. **Influencer and reseller agreement templates.**
   - Research standard affiliate, distribution, and OEM licensing terms for hardware accessories in the medical/radiology market.

---

## Bibliography

[1] OpenCommunityLicence (2026). "OpenCommunityLicence/OpenCommunityLicence — OCL v1.1 Repository". GitHub. https://github.com/OpenCommunityLicence/OpenCommunityLicence (Retrieved: 2026-07-15)

[2] Prusa Research (2026). "Open Community License v1.1: The New Plugin System, More Examples, and Your Questions Answered". Original Prusa 3D Printers Blog. https://blog.prusa3d.com/open-community-license-v1-1-the-new-plugin-system-more-examples-and-your-questions-answered_137202/ (Retrieved: 2026-07-15)

[3] Mitchell, K. (2026). "Open Community License v1". /dev/lawyer. https://writing.kemitchell.com/2026/03/16/Open-Community-License-1 (Retrieved: 2026-07-15)

[4] Open Source Hardware Association. "Open Source Hardware Definition 1.0". OSHWA. https://oshwa.org/definition/ (Retrieved: 2026-07-15)

[5] PolyForm Project. "PolyForm Noncommercial License 1.0.0". https://polyformproject.org/licenses/noncommercial/1.0.0 (Retrieved: 2026-07-15)

[6] Fair-code. "Fair-code — Software Model and Principles". https://faircode.io/ (Retrieved: 2026-07-15)

[7] Commons Clause. "Commons Clause License Condition v1.0". https://commonsclause.com/ (Retrieved: 2026-07-15)

[8] CERN. "CERN Open Hardware Licence Version 2". https://cern-ohl.web.cern.ch/ (Retrieved: 2026-07-15)

[9] TAPR. "The TAPR Open Hardware License". https://tapr.org/the-tapr-open-hardware-license/ (Retrieved: 2026-07-15)

[10] Open Source Stack Exchange. "How can a project be relicensed?". https://opensource.stackexchange.com/questions/33/how-can-a-project-be-relicensed (Retrieved: 2026-07-15)

[11] Montón, M. & Salazar, X. (2020). "On licenses for [Open] Hardware". arXiv:2010.09039 [cs.OH]. https://arxiv.org/abs/2010.09039 (Retrieved: 2026-07-15)

[12] TAPR. "TAPR Noncommercial Hardware License v1.0". https://files.tapr.org/OHL/TAPR_Noncommercial_Hardware_License_v1.0.txt (Retrieved: 2026-07-15)

[13] The Turing Way. "Open Hardware Licenses". https://book.the-turing-way.org/reproducible-research/licensing/licensing-hardware/ (Retrieved: 2026-07-15)

[14] Open Source Initiative. "The MIT License". https://opensource.org/license/MIT (Retrieved: 2026-07-15)

[15] 4Additive (2026). "Prusa's OCL: open source or hidden barrier?". https://www.4additive.com/en/prusa-ocl-open-source-hidden-barrier/ (Retrieved: 2026-07-15)

[16] Tom's Hardware (2025). "Prusa Research introduces the Open Community License to protect open source 3D printing hardware". https://www.tomshardware.com/3d-printing/prusa-research-introduces-the-open-community-license-to-protect-open-source-3d-printing-hardware-new-rules-aimed-at-addressing-industry-abuses (Retrieved: 2026-07-15)

---

## Appendix: Methodology

### Research Process

This report followed the Deep Research 8-phase pipeline in standard mode.

- **Phase 1 (SCOPE):** Defined the question around combined hardware-software licenses that allow community access but restrict commercial sale, using 3D printing as the primary comparable industry.
- **Phase 2 (PLAN):** Identified primary sources (OCL repository, OSHWA, license steward pages), legal commentary (/dev/lawyer, 4Additive, Adafruit), and foundational references (CERN OHL, TAPR OHL, PolyForm, Commons Clause, arXiv survey).
- **Phase 3 (RETRIEVE):** Executed parallel web searches and direct fetches of primary license texts and analyses. Sources were registered in a citation manager and evidence was extracted.
- **Phase 4 (TRIANGULATE):** Cross-checked claims about OCL terms against the repository, Prusa's blog, and independent legal analysis. Verified that CERN OHL and TAPR OHL permit commercial sale against SPDX and steward pages.
- **Phase 5 (SYNTHESIZE):** Identified patterns, generated insights about the "source-available hardware" gap, and assessed implications for RadKeys.
- **Phase 6 (CRITIQUE):** Reviewed contradictions (OCL ambiguities, community rejection, TAPR NCL deprecation) and limitations (no case law, Brazilian law gap).
- **Phase 7 (REFINE):** Added recommendations for authorship verification, commercial licensing page, trademark registration, and legal review.
- **Phase 8 (PACKAGE):** Assembled this report with progressive section generation.

### Sources Consulted

**Total Sources:** 16 registered.

**Source Types:**
- License steward documentation: 8
- Project/vendor blog posts: 2
- Legal analysis / commentary: 2
- Community Q&A: 1
- Academic paper: 1
- Standard-setting organization: 1
- News / trade publication: 1

**Temporal Coverage:** 2007 (TAPR NCL) to June 2026 (Prusa OCL v1.1 announcement).

### Verification Approach

**Triangulation:** Major claims about license terms were verified against primary texts rather than secondary summaries. For example, OCL's sale restriction was confirmed by reading the repository's summary and the Prusa blog, not only news articles.

**Credibility Assessment:** Primary sources (license stewards, OSHWA, Prusa) were weighted highest. Legal commentary was treated as interpretation, not fact. Community criticism was included to surface reputational risks.

**Quality Control:** Claims unsupported by evidence were excluded. Where direct fetch failed (Adafruit blog, 403), the source was cited only for its existence in search results and not used as primary evidence.

### Claims-Evidence Table

| Claim ID | Major Claim | Evidence Type | Supporting Sources | Confidence |
|----------|-------------|---------------|-------------------|------------|
| C1 | OCL v1.1 covers both hardware and software | Primary text | [1] | High |
| C2 | OCL restricts commercial sale/replication without separate license | Primary text | [1][2] | High |
| C3 | OCL is not OSI/OSHWA open source | Definition + analysis | [3][4][15] | High |
| C4 | PolyForm/Commons Clause are software-only | Primary text | [5][7] | High |
| C5 | CERN OHL and TAPR OHL allow commercial sale | Primary text + SPDX | [8][9] | High |
| C6 | Relicensing MIT requires contributor consent | Community consensus | [10] | High |
| C7 | OCL has drafting ambiguities | Legal commentary | [3] | Medium |
| C8 | No case law interprets OCL | Known gap | — | Low (acknowledged) |

---

## Report Metadata

**Research Mode:** Standard  
**Total Sources:** 16  
**Approximate Word Count:** 4,200  
**Research Duration:** ~25 minutes  
**Generated:** 2026-07-15  
**Validation Status:** Self-reviewed; no automated validation run. Recommend running `validate_report.py` and `verify_citations.py` before publication.
