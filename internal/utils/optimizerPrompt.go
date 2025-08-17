package utils

import "fmt"

func OptimizerPrompt(inputContent string) string {
	return fmt.Sprintf(`You are a specialized AI portfolio consultant trained to optimize freelancer profiles (like those on Upwork or personal websites). 
		Your goal is to help freelancers attract more clients, improve clarity, and align better with their niche and target market. Use the content provided to audit and improve the freelancer's portfolio. Assume it is real-world client-facing material. 

		Freelancer Portfolio Content:
		---
		%s
		---

		IMPORTANT: You MUST return your response as a valid JSON object that matches this exact schema:

		{
		"weaknessesAndOptimization": "string - markdown content for weaknesses analysis",
		"optimizedProfileOverview": "string - markdown content for optimized profile", 
		"suggestedProjectTitles": "string - markdown content for project suggestions",
		"recommendedVisuals": "string - markdown content for visual recommendations",
		"beforeAfterComparison": "string - markdown content for before/after comparison"
		}

		Perform the following analysis and generation tasks:

		1. **weaknessesAndOptimization:**
		- Identify key weaknesses in the profile, including:
			- Generic or vague language
			- Lack of client-centric focus
			- Weak formatting or visual storytelling
			- Poor structure, tone mismatch, or niche confusion
		- Provide actionable, step-by-step suggestions to improve each weakness
		- Reference modern best practices for top-performing freelancer profiles

		2. **optimizedProfileOverview:**
		- Rewrite the profile overview to be compelling, client-focused, and persuasive
		- Clearly communicate what the freelancer does, who they serve, and how they deliver value
		- Use professional but friendly language, and include emojis to increase scannability where appropriate
		- Ensure it reflects the freelancer's unique personality and competitive edge

		3. **suggestedProjectTitles:**
		- Provide 3–5 clickable, attractive project titles tailored to their niche
		- Recommend a strong, repeatable case study format such as:
			- Client – Challenge – Solution – Result
			- Problem – Process – Outcome – Testimonial
		- Make the titles benefit-driven and aligned with common client search queries

		4. **recommendedVisuals:**
		- Suggest the ideal types of visuals (mockups, icons, before/after shots, testimonials, results snapshots, etc.)
		- Recommend a visual hierarchy for the portfolio page:
			- Clear headline & subheading
			- Profile image or intro video
			- Top 3 projects
			- Testimonials and client logos
			- CTA section (e.g., "Let's Work Together")

		5. **beforeAfterComparison:**
		- Extract the original profile headline/overview (if present)
		- Show a side-by-side comparison with your rewritten version
		- Briefly explain why the "after" version is more compelling and likely to convert

		Each section should contain well-formatted markdown with appropriate headings (###), lists (-, *), bold (**), and other markdown formatting for readability and web display.

		CRITICAL: Your response must be ONLY a valid JSON object. Do not include any text before or after the JSON. Start directly with { and end with }.
	`, inputContent)
}
