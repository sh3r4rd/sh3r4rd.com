"""Recruiter data extraction using the OpenAI ChatGPT API.

Replaces regex-based extraction with a single structured LLM call that
returns company, job_title, name, and phone as an ExtractionResult.
The model is expected via the ``OPENAI_API_KEY`` environment variable.
Fields that cannot be determined are returned as ``None``.
"""

from __future__ import annotations

import json
import logging
import os
from dataclasses import dataclass
from typing import Optional

import openai

logger = logging.getLogger(__name__)

_MODEL = "gpt-4o-mini"

# Initialised once at module load so Lambda warm starts reuse the same client.
# Raises EnvironmentError with a clear message when the variable is absent.
_api_key = os.environ.get("OPENAI_API_KEY")
if not _api_key:
    raise EnvironmentError(
        "OPENAI_API_KEY environment variable is required but not set."
    )
_client = openai.OpenAI(api_key=_api_key)

_SYSTEM_PROMPT = (
    "You are a structured data extraction assistant. "
    "Extract recruiter information from the provided email and return "
    "a JSON object with exactly four keys: "
    "\"company\" (the recruiter's employer), "
    "\"job_title\" (the role being hired for), "
    "\"name\" (the recruiter's full name), and "
    "\"phone\" (the recruiter's phone number). "
    "Set a key to null when the information is not present in the email. "
    "Return only valid JSON — no markdown, no explanation."
)

_USER_TEMPLATE = """\
From: {from_header}
Sender email: {sender_email}

--- Email body ---
{body}
"""


@dataclass
class ExtractionResult:
    """Structured recruiter data extracted from an email."""

    company: Optional[str] = None
    job_title: Optional[str] = None
    name: Optional[str] = None
    phone: Optional[str] = None


def extract(
    email_body: str,
    from_header: Optional[str] = None,
    sender_email: Optional[str] = None,
) -> ExtractionResult:
    """Extract recruiter fields from *email_body* using ChatGPT.

    Args:
        email_body: Plain-text body of the recruiter email.
        from_header: Raw ``From:`` header value (e.g. ``"Alice <alice@corp.com>"``).
        sender_email: Sender's email address, used as a fallback for company
            inference when the body text is ambiguous.

    Returns:
        An :class:`ExtractionResult` with ``None`` for any field that could not
        be determined from the email content.

    Raises:
        openai.OpenAIError: If the API call fails.
    """
    user_message = _USER_TEMPLATE.format(
        from_header=from_header or "Not provided",
        sender_email=sender_email or "Not provided",
        body=email_body,
    )

    response = _client.chat.completions.create(
        model=_MODEL,
        messages=[
            {"role": "system", "content": _SYSTEM_PROMPT},
            {"role": "user", "content": user_message},
        ],
        response_format={"type": "json_object"},
        temperature=0,
    )

    raw = response.choices[0].message.content or "{}"
    data: dict = json.loads(raw)

    # Normalise empty strings to None: both null and "" from the model mean
    # "not found", matching the contract that unextracted fields are None.
    return ExtractionResult(
        company=data.get("company") or None,
        job_title=data.get("job_title") or None,
        name=data.get("name") or None,
        phone=data.get("phone") or None,
    )
