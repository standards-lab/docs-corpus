#!/usr/bin/env -S python3 -I
"""Prints the trait and feature placement of every ~/docs sample as a JSON array, sorted by path.

With --facts, each element also has the probes' findings. Writes nothing outside a temporary
directory.

See docs/coverage.md
"""

import collections
import email
import email.header
import email.policy
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import unicodedata
import zipfile

ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
DOCS = os.environ.get("DOCS_REFERENCE") or os.path.expanduser("~/docs")
RUNS = os.environ.get("RUNS") or os.path.expanduser("~/runs/docproc/final/runs")

SCRIPT_THRESHOLD = {"latin": 60, "han": 10, "kana": 10, "hangul": 10}
SENTENCE = 20
MIB = 1024 * 1024
OOXML_KINDS = ("docx", "xlsx", "pptx", "docm", "dotm", "xlsm", "xltm", "pptm", "dotx", "xltx", "potx", "ppsx")
LEGACY_TARGET = {".doc": "docx", ".rtf": "docx", ".wpd": "docx", ".xls": "xlsx", ".ppt": "pptx"}
LATIN_GUESSES = {"WINDOWS-1252", "WINDOWS-1250", "WINDOWS-1254", "WINDOWS-1257", "ISO-8859-1", "ISO-8859-2",
                 "ISO-8859-3", "ISO-8859-4", "ISO-8859-9", "ISO-8859-15", "ISO-8859-16"}
UTF8 = {"utf-8", "utf8", "us-ascii", "ascii", ""}


def run(args, data=None, timeout=600):
    try:
        out = subprocess.run(args, input=data, capture_output=True, timeout=timeout)
    except FileNotFoundError:
        sys.exit(f"place-traits: {args[0]} is not on PATH; install it")
    except (OSError, subprocess.TimeoutExpired):
        return b""
    return out.stdout


def script_of(ch):
    if not ch.isalpha():
        return None
    try:
        name = unicodedata.name(ch)
    except ValueError:
        return None
    for prefix, script in (
        ("LATIN", "latin"), ("CYRILLIC", "cyrillic"), ("GREEK", "greek"), ("ARABIC", "arabic"),
        ("HEBREW", "hebrew"), ("DEVANAGARI", "devanagari"), ("HIRAGANA", "kana"),
        ("KATAKANA", "kana"), ("HANGUL", "hangul"), ("CJK", "han"),
    ):
        if name.startswith(prefix):
            return script
    return None


def scripts_in(text):
    counts = collections.Counter(s for s in map(script_of, text) if s)
    found = {s for s, n in counts.items() if n >= SCRIPT_THRESHOLD.get(s, SENTENCE)}
    traits = {f"trait:script.{s}" for s in found if s != "latin"}
    systems = {("japanese" if s in ("han", "kana") and "kana" in found else s) for s in found}
    if len(systems) >= 2:
        traits.add("trait:script.mixed")
    return traits, dict(counts)


class Run:
    def __init__(self, run_id):
        self.content = os.path.join(os.path.dirname(RUNS), "content")
        self.events = collections.defaultdict(list)
        with open(os.path.join(RUNS, run_id, "manifest.json")) as f:
            root = json.load(f)["inputs"][0] + "/"
        with open(os.path.join(RUNS, run_id, "record.jsonl")) as f:
            for line in f:
                e = json.loads(line)
                path = e.get("input", "")
                if path.startswith(root):
                    self.events[path[len(root):]].append(e)

    def blob(self, sha):
        try:
            with open(os.path.join(self.content, sha[:2], sha), "rb") as f:
                return f.read()
        except OSError:
            return b""

    def outputs(self, path):
        return [e for e in self.events.get(path, []) if e.get("event") == "output"]

    def text(self, path):
        parts = []
        for e in self.outputs(path):
            out = e["output"]
            if out.get("kind") not in ("text", "markdown", "record"):
                continue
            data = self.blob(out.get("sha256", ""))
            encoding = (out.get("meta") or {}).get("encoding") or "utf-8"
            try:
                body = data.decode(encoding, errors="replace")
            except LookupError:
                body = data.decode("utf-8", errors="replace")
            if out["kind"] == "record":
                for line in body.splitlines():
                    try:
                        parts.extend(strings(json.loads(line), ("markdown", "subject", "text", "title")))
                    except ValueError:
                        pass
            else:
                parts.append(body)
        return "\n".join(parts)

    def records(self, path, media):
        rows = []
        for e in self.outputs(path):
            out = e["output"]
            if out.get("kind") == "record" and media in out.get("media_type", ""):
                for line in self.blob(out.get("sha256", "")).decode("utf-8", errors="replace").splitlines():
                    try:
                        rows.append((out.get("name", ""), json.loads(line)))
                    except ValueError:
                        pass
        return rows


def strings(value, keys):
    if isinstance(value, dict):
        for k, v in value.items():
            if k in keys and isinstance(v, str):
                yield v
            else:
                yield from strings(v, keys)
    elif isinstance(value, list):
        for v in value:
            yield from strings(v, keys)


def zip_texts(path):
    try:
        with zipfile.ZipFile(path) as z:
            return {n: z.read(n) for n in z.namelist()}
    except (zipfile.BadZipFile, OSError, RuntimeError, NotImplementedError):
        return {}


def ooxml(path):
    files = zip_texts(path)
    names = list(files)
    xml = {n: files[n].decode("utf-8", errors="replace") for n in names if n.endswith((".xml", ".rels"))}
    body = "\n".join(v for n, v in xml.items() if not n.endswith(".rels"))
    rels = "\n".join(v for n, v in xml.items() if n.endswith(".rels"))
    sheets = [v for n, v in xml.items() if re.match(r"xl/worksheets/sheet\d+\.xml$", n)]
    slides = [v for n, v in xml.items() if re.match(r"ppt/slides/slide\d+\.xml$", n)]
    word = [v for n, v in xml.items() if re.match(r"word/(document|header\d*|footer\d*|footnotes|endnotes)\.xml$", n)]
    features = set()
    facts = {"parts": len(names)}
    tables = sum(len(re.findall(r"<w:tbl[ >]", v)) for v in word) + sum(len(re.findall(r"<a:tbl[ >]", v)) for v in slides)
    if tables:
        features.add("feature:tables")
    if any(re.search(r"<w:tc[ >](?:(?!</w:tc>).)*<w:tbl[ >]", v, re.S) for v in word):
        features.add("feature:nested-tables")
    if any(re.search(r"<w:gridSpan |<w:vMerge", v) for v in word) or any(
        re.search(r'gridSpan="|rowSpan="|hMerge="1"|vMerge="1"', v) for v in slides
    ) or any("<mergeCell " in v for v in sheets):
        features.add("feature:merged-cells")
    if any(re.search(r"(charts|chartsheets)/(chart|chartEx)\d*\.xml$", n) for n in names):
        features.add("feature:charts")
    if any(re.match(r"(word/comments\.xml|ppt/comments/|ppt/modernComment|word/commentsExtended)", n) for n in names):
        if re.search(r"<w:comment |<p:cm |<p188:cm ", body):
            features.add("feature:comments")
    if any(re.match(r"xl/comments\d*\.xml$", n) for n in names) or any(re.match(r"xl/threadedComments/", n) for n in names):
        features.add("feature:cell-notes")
    if any(re.search(r"<w:(ins|del) ", v) for v in word):
        features.add("feature:tracked-changes")
    notes = xml.get("word/footnotes.xml", "") + xml.get("word/endnotes.xml", "")
    if re.search(r'<w:(footnote|endnote)(?![^>]*w:type="(separator|continuationSeparator|continuationNotice)")[^>]*w:id=', notes):
        features.add("feature:footnotes")
    if any(re.match(r"ppt/notesSlides/notesSlide\d+\.xml$", n) and re.search(r"<a:t>[^<]*\S", xml[n]) for n in xml):
        features.add("feature:speaker-notes")
    if re.search(r"<w:txbxContent|<v:textbox", body) or any(
        re.search(r"<p:sp>(?:(?!</p:sp>).)*<p:txBody>(?:(?!</p:sp>).)*<a:t>", re.sub(r"<p:sp>(?:(?!</p:sp>).)*<p:ph[ />](?:(?!</p:sp>).)*</p:sp>", "", v, flags=re.S), re.S)
        for v in slides
    ) or any(re.match(r"xl/drawings/drawing\d+\.xml$", n) and re.search(r"<xdr:txBody>(?:(?!</xdr:txBody>).)*<a:t>", xml[n], re.S) for n in xml):
        features.add("feature:text-boxes")
    if "<m:oMath" in body:
        features.add("feature:equations")
    if any(re.search(r"/diagrams/data\d*\.xml$", n) for n in names):
        features.add("feature:smartart")
    if any(re.search(r"<f[ >]", v) for v in sheets):
        features.add("feature:formulas")
    if sheets and xlsx_dates(xml):
        features.add("feature:dates")
    if re.search(r'Type="[^"]*/hyperlink"[^>]*TargetMode="External"|TargetMode="External"[^>]*Type="[^"]*/hyperlink"', rels):
        features.add("feature:hyperlinks")
    media = [n for n in names if re.search(r"/media/", n)]
    if media:
        features.add("feature:pictures")
    if any(re.search(r'<a:srcRect [^>]*[lrtb]="[1-9]', v) for v in xml.values()) or slide_offpage(xml):
        features.add("feature:picture-layout")
    if any(re.search(r"/embeddings/", n) for n in names) or "<o:OLEObject" in body or "<p:oleObj" in body:
        features.add("feature:embedded-objects")
    facts.update(tables=tables, media=len(media))
    return features, facts


def xlsx_dates(xml):
    styles = xml.get("xl/styles.xml", "")
    date_formats = {str(i) for i in list(range(14, 23)) + list(range(45, 48))}
    for m in re.finditer(r'<numFmt numFmtId="(\d+)" formatCode="([^"]*)"', styles):
        code = re.sub(r'&quot;.*?&quot;|"[^"]*"|\[[^\]]*\]|\\.', "", m.group(2)).lower()
        if re.search(r"[dmyhs]", code):
            date_formats.add(m.group(1))
    xfs = re.search(r"<cellXfs[^>]*>(.*?)</cellXfs>", styles, re.S)
    styled = set()
    for i, m in enumerate(re.finditer(r"<xf ([^>]*)", xfs.group(1) if xfs else "")):
        fmt = re.search(r'numFmtId="(\d+)"', m.group(1))
        if fmt and fmt.group(1) in date_formats:
            styled.add(str(i))
    sheets = "\n".join(v for n, v in xml.items() if n.startswith("xl/worksheets/sheet"))
    for m in re.finditer(r'<c ([^>]*)>(.*?)</c>', sheets, re.S):
        style = re.search(r'\bs="(\d+)"', m.group(1))
        if style and style.group(1) in styled and "<v>" in m.group(2) and 't="s"' not in m.group(1):
            return True
    return False


def slide_offpage(xml):
    pres = xml.get("ppt/presentation.xml", "")
    m = re.search(r'<p:sldSz cx="(\d+)" cy="(\d+)"', pres)
    if not m:
        return False
    cx, cy = int(m.group(1)), int(m.group(2))
    for n, v in xml.items():
        if not re.match(r"ppt/slides/slide\d+\.xml$", n):
            continue
        for pic in re.findall(r"<p:pic>.*?</p:pic>", v, re.S):
            off = re.search(r'<a:off x="(-?\d+)" y="(-?\d+)"/><a:ext cx="(\d+)" cy="(\d+)"', pic)
            if off:
                x, y, w, h = map(int, off.groups())
                if x < 0 or y < 0 or x + w > cx or y + h > cy:
                    return True
    return False


def converted(path, tmp):
    ext = os.path.splitext(path)[1].lower()
    target = LEGACY_TARGET.get(ext)
    if not target:
        return None
    out = os.path.join(tmp, str(abs(hash(path))))
    os.makedirs(out, exist_ok=True)
    run(["soffice", f"-env:UserInstallation=file://{tmp}/profile", "--headless", "--convert-to", target, "--outdir", out, path], timeout=300)
    for name in os.listdir(out):
        return os.path.join(out, name)
    return None


def pdf(path, rel, runrec):
    facts = {}
    traits, features = set(), set()
    info = run(["pdfinfo", path]).decode("utf-8", errors="replace")
    m = re.search(r"^Pages:\s+(\d+)", info, re.M)
    pages = int(m.group(1)) if m else 0
    outs = runrec.outputs(rel)
    top = [e for e in outs if e["part"].get("depth", 0) == 0]
    text_pages = {e["output"]["meta"].get("page") for e in top if e["output"].get("kind") == "text" and e["output"].get("name", "").startswith("page ")}
    image_pages = [e for e in top if e["output"].get("kind") == "image" and e["output"].get("name", "").startswith("page ")]
    dpis = {(e["output"].get("meta") or {}).get("dpi") for e in image_pages}
    rendered = len(image_pages) or pages
    facts.update(pages=pages, rendered_pages=len(image_pages), text_pages=len(text_pages), dpis=sorted(d for d in dpis if d))
    if rendered:
        if not text_pages:
            traits.add("trait:text-layer.none")
        elif len(text_pages) < rendered:
            traits.add("trait:text-layer.partial")
    if len(dpis) > 1:
        traits.add("trait:scale.oversized-page")
    if max(pages, rendered) >= 100:
        traits.add("trait:scale.many-pages")
    images = run(["pdfimages", "-list", path]).decode("utf-8", errors="replace").splitlines()[2:]
    image_page_set = {int(l.split()[0]) for l in images if l.split() and l.split()[0].isdigit() and int(l.split()[3]) >= 500}
    facts["pages_with_large_images"] = len(image_page_set)
    ppi = [int(l.split()[12]) for l in images if len(l.split()) > 12 and l.split()[12].isdigit() and int(l.split()[3]) >= 500]
    facts["large_image_ppi"] = sorted(set(ppi))[:3]
    fonts = run(["pdffonts", path]).decode("utf-8", errors="replace").splitlines()[2:]
    facts["fonts"] = len(fonts)
    facts["fonts_without_unicode"] = sum(1 for l in fonts if re.search(r"\s(yes|no)\s+(yes|no)\s+no\s+\d+\s+\d+\s*$", l))
    raw = run(["qpdf", "--qdf", "--object-streams=disable", "--decrypt", path, "-"])
    if not raw:
        with open(path, "rb") as f:
            raw = f.read()
    facts["invisible_text_runs"] = len(re.findall(rb"(?m)(^|[^0-9.])3\s+Tr\b", raw))
    if text_pages and image_page_set and facts["invisible_text_runs"] and len(image_page_set) >= max(1, len(text_pages) // 2):
        traits.add("trait:text-layer.ocr")
    if re.search(rb"/AcroForm\s*(<<|\d+ 0 R)", raw) and re.search(rb"/FT\s*/(Tx|Btn|Ch)", raw):
        features.add("feature:forms")
    if re.search(rb"/FT\s*/Sig", raw):
        features.add("feature:signature-fields")
    if re.search(rb"/Collection\s*(<<|\d+ 0 R)", raw):
        features.add("feature:portfolio")
    return traits, features, facts


def audio_video(path):
    data = run(["ffprobe", "-v", "error", "-show_streams", "-show_format", "-of", "json", path])
    try:
        probe = json.loads(data or b"{}")
    except ValueError:
        probe = {}
    streams = probe.get("streams", [])
    audio = [s for s in streams if s.get("codec_type") == "audio"]
    video = [s for s in streams if s.get("codec_type") == "video" and not (s.get("disposition") or {}).get("attached_pic")]
    pics = [s for s in streams if (s.get("disposition") or {}).get("attached_pic")]
    duration = float((probe.get("format") or {}).get("duration") or 0)
    rates = sorted({int(s.get("sample_rate", 0)) for s in audio})
    facts = {"duration": round(duration, 1), "audio_streams": len(audio), "video_streams": len(video),
             "attached_pictures": len(pics), "sample_rates": rates,
             "channels": sorted({s.get("channels") for s in audio if s.get("channels")}),
             "subtitle_streams": sum(1 for s in streams if s.get("codec_type") == "subtitle")}
    traits, features = set(), set()
    if video and not audio:
        traits.add("trait:audio.none")
    if audio and max(rates) <= 8000:
        traits.add("trait:audio.narrowband")
    if audio and duration > 30:
        traits.add("trait:audio.multi-window")
    if duration >= 1800:
        traits.add("trait:scale.long-media")
    if pics:
        features.add("feature:attached-picture")
    if len(audio) > 1 or len(video) > 1:
        features.add("feature:multiple-streams")
    return traits, features, facts


def image(path):
    out = run(["identify", "-format", "%m|%w|%h|%A|%[orientation]|%[opaque]\n", path]).decode("utf-8", errors="replace")
    frames = [l.split("|") for l in out.splitlines() if l.count("|") == 5]
    facts = {"frames": len(frames)}
    traits, features = set(), set()
    if not frames:
        return traits, features, facts
    fmt, w, h, alpha, orient, opaque = frames[0]
    facts.update(format=fmt, width=int(w), height=int(h), alpha=alpha, orientation=orient, opaque=opaque)
    if len(frames) > 1 and fmt == "TIFF":
        features.add("feature:multi-page-image")
    if len(frames) > 1 and fmt in ("GIF", "WEBP", "AVIF"):
        features.add("feature:animation")
    if orient not in ("", "Undefined", "TopLeft"):
        features.add("feature:orientation")
    if alpha not in ("Undefined", "False", "") and opaque.lower() == "false":
        features.add("feature:transparency")
    return traits, features, facts


def charset(data):
    if data.startswith((b"\xef\xbb\xbf", b"\xff\xfe", b"\xfe\xff")):
        return None, "bom"
    try:
        data.decode("utf-8")
        return None, "utf-8"
    except UnicodeDecodeError:
        pass
    for enc in ("utf-16-be", "utf-16-le"):
        if b"\0" in data and len(data) % 2 == 0:
            try:
                decoded = data.decode(enc)
            except UnicodeDecodeError:
                continue
            if sum(c.isprintable() or c.isspace() for c in decoded) >= 0.98 * len(decoded):
                return "trait:charset.utf-16-no-bom", enc.upper()
    guess = run(["uchardet"], data=data).decode().strip().upper()
    if guess.startswith(("UTF-16", "UTF-32")):
        return "trait:charset.utf-16-no-bom", guess
    if guess in LATIN_GUESSES:
        try:
            data.decode("cp1252")
            return "trait:charset.windows-1252", guess
        except UnicodeDecodeError:
            pass
    return "trait:charset.legacy-other", guess


def decode_text(data, guess):
    if data.startswith(b"\xef\xbb\xbf"):
        return data[3:].decode("utf-8", errors="replace")
    if data.startswith((b"\xff\xfe", b"\xfe\xff")):
        return data.decode("utf-16", errors="replace")
    if guess.startswith("UTF-16"):
        return data.decode("utf-16-be" if data[:1] == b"\0" else "utf-16-le", errors="replace")
    try:
        return data.decode(guess if guess not in ("utf-8", "bom", "") else "utf-8", errors="replace")
    except LookupError:
        return data.decode("utf-8", errors="replace")


def message(data):
    msg = email.message_from_bytes(data, policy=email.policy.compat32)
    features, traits = set(), set()
    texts = []
    facts = {"parts": 0}
    if re.search(rb"=\?[^?]+\?[bBqQ]\?", data.split(b"\n\n", 1)[0]):
        traits.add("trait:charset.declared")
    bodies = set()
    for part in msg.walk():
        facts["parts"] += 1
        ctype = part.get_content_type()
        cs = (part.get_content_charset() or "").lower()
        if cs not in UTF8:
            traits.add("trait:charset.declared")
        disp = (part.get("Content-Disposition") or "").lower()
        if ctype == "message/rfc822" or (ctype.startswith("message/") and part is not msg and ctype != "message/delivery-status"):
            features.add("feature:embedded-messages")
        if ctype == "application/ms-tnef" or (part.get_filename() or "").lower() == "winmail.dat":
            features.add("feature:tnef-part")
        if ctype in ("application/pkcs7-mime", "application/x-pkcs7-mime"):
            features.add("feature:smime-opaque")
        if ctype == "multipart/signed" and "pkcs7" in (part.get_param("protocol") or ""):
            features.add("feature:smime-signed")
        if (part.get_filename() or "").lower() == "message.rpmsg":
            features.add("feature:rights-managed")
        if part.get("Content-ID") and ctype.startswith("image/"):
            features.add("feature:inline-images")
        if "attachment" in disp or (part.get_filename() and not ctype.startswith("multipart/") and part.get("Content-ID") is None):
            features.add("feature:attachments")
        if ctype in ("text/plain", "text/html") and "attachment" not in disp:
            bodies.add(ctype)
            payload = part.get_payload(decode=True) or b""
            try:
                text = payload.decode(cs or "utf-8", errors="replace")
            except LookupError:
                text = payload.decode("utf-8", errors="replace")
            texts.append(text)
            if ctype == "text/html":
                if re.search(r'<img[^>]+src=["\']?(data:|cid:)', text, re.I):
                    features.add("feature:inline-images")
                if re.search(r'<a[^>]+href=["\']?https?://', text, re.I):
                    features.add("feature:hyperlinks")
            elif re.search(r"https?://\S+", text):
                features.add("feature:hyperlinks")
    if bodies == {"text/html"}:
        features.add("feature:html-body")
    if msg.get("Subject"):
        texts.append(str(email.header.make_header(email.header.decode_header(msg["Subject"]))))
    return traits, features, "\n".join(texts), facts


def mail_records(rows):
    traits, features = set(), set()
    facts = collections.Counter()
    ids, folders_with_messages = set(), set()
    refs = []
    top = 0
    for name, r in rows:
        if not name.startswith("message") or not ({"subject", "sender", "transport_headers"} & set(r)):
            continue
        facts["messages"] += 1
        top += "id" in r
        if r.get("folder") is not None:
            folders_with_messages.add(r.get("folder"))
        cls = (r.get("class") or "IPM.Note")
        if not cls.upper().startswith(("IPM.NOTE", "REPORT.IPM.NOTE")) or cls.upper().startswith("IPM.NOTE.RULES"):
            if re.match(r"IPM\.(Contact|Appointment|Task|StickyNote|Activity|Schedule|DistList)", cls, re.I):
                features.add("feature:non-note-items")
        if r.get("message_id"):
            ids.add(r["message_id"])
        refs.append((r.get("in_reply_to") or "") + " " + (r.get("references") or ""))
        hdr = r.get("transport_headers") or ""
        if re.search(r"=\?[^?]+\?[bBqQ]\?", hdr) or re.search(r"charset=\"?(?!utf-8|us-ascii)[\w-]+", hdr, re.I):
            traits.add("trait:charset.declared")
        for a in r.get("attachments") or []:
            method = a.get("method", "")
            mime = a.get("mime_type", "")
            if method == "embedded_message" or mime == "message/rfc822":
                features.add("feature:embedded-messages")
            elif a.get("inline"):
                features.add("feature:inline-images")
            elif method.startswith("by_ref"):
                features.add("feature:attachment-by-reference")
            else:
                features.add("feature:attachments")
            if mime == "application/ms-tnef":
                features.add("feature:tnef-part")
            if "pkcs7-mime" in mime:
                features.add("feature:smime-opaque")
            if (a.get("name") or a.get("filename") or "").lower() == "message.rpmsg":
                features.add("feature:rights-managed")
        md = r.get("markdown") or ""
        if re.search(r"\]\(https?://|<https?://|https?://", md):
            features.add("feature:hyperlinks")
    if top > 1 and any(any(i and i in ref for i in ids) for ref in refs):
        features.add("feature:threads")
    if len(folders_with_messages) > 1:
        features.add("feature:folders")
    if facts["messages"] >= 1000:
        traits.add("trait:scale.many-messages")
    facts["folders_with_messages"] = len(folders_with_messages)
    return traits, features, dict(facts)


def msg_bodies(path):
    with open(path, "rb") as f:
        data = f.read()
    names = re.findall(rb"_\x00_\x00s\x00u\x00b\x00s\x00t\x00g\x001\x00\.\x000\x00_\x00((?:[0-9A-F]\x00){4})", data)
    return sorted({n.decode("utf-16-le") for n in names} & {"1000", "1009", "1013"})


def mbox(path):
    with open(path, "rb") as f:
        data = f.read()
    chunks = re.split(rb"\n(?=From [^\n]*\n)", data)
    ids, refs = set(), []
    traits, features = set(), set()
    texts = []
    for c in chunks:
        msg = email.message_from_bytes(c, policy=email.policy.compat32)
        if msg.get("Message-ID"):
            ids.add(msg["Message-ID"].strip())
        refs.append((msg.get("In-Reply-To") or "") + " " + (msg.get("References") or ""))
        t, f, text, _ = message(c)
        traits |= t
        features |= f
        texts.append(text)
    if any(any(i in r for i in ids) for r in refs):
        features.add("feature:threads")
    if len(chunks) >= 1000:
        traits.add("trait:scale.many-messages")
    return traits, features, "\n".join(texts), {"messages": len(chunks)}


def onenote(rows, outcomes):
    traits, features = set(), set()
    facts = collections.Counter()
    if "encrypted" in outcomes:
        features.add("feature:protected-section")
    for name, r in rows:
        if name.startswith("section"):
            facts["sections"] += 1
            if len(r.get("path") or []) > 1:
                features.add("feature:section-groups")
        if not name.startswith("page"):
            continue
        facts["pages"] += 1
        md = r.get("markdown") or ""
        if re.search(r"^\|.*\|\s*$", md, re.M):
            features.add("feature:tables")
        if re.search(r"^\s*- \[[ x]\] |\[(To Do|Important|Question|Idea|Critical)\]", md, re.M):
            features.add("feature:note-tags")
        if re.search(r"\]\(https?://", md):
            features.add("feature:hyperlinks")
        if re.search(r"\$\$[^$]+\$\$", md):
            features.add("feature:equations")
        if r.get("ink_alternatives"):
            traits.add("trait:scan.handwriting")
        if r.get("revisions"):
            features.add("feature:page-revisions")
        for m in r.get("media") or []:
            kind = m.get("kind")
            if kind == "image":
                features.add("feature:pictures")
            elif kind == "ink":
                features.add("feature:ink")
            elif kind == "embedded_file":
                features.add("feature:embedded-objects")
            if m.get("ocr") or m.get("text"):
                features.add("feature:image-ocr")
    return traits, features, dict(facts)


def odf_text(path):
    files = zip_texts(path)
    xml = files.get("content.xml", b"").decode("utf-8", errors="replace")
    return re.sub(r"<[^>]+>", " ", xml)


def archive_entries(path):
    out = run(["bsdtar", "-tf", path]).decode("utf-8", errors="replace").splitlines()
    return [l for l in out if l and not l.endswith("/")]


def probe(sample, entry, runrec, tmp):
    rel = sample["path"]
    path = os.path.join(DOCS, rel)
    library = set(sample.get("library", []))
    ext = os.path.splitext(rel)[1].lower()
    traits, features, facts = set(), set(), {}
    text = None
    size = os.path.getsize(path)
    if size >= 100 * MIB:
        traits.add("trait:scale.large-file")
    with open(path, "rb") as f:
        head = f.read(8)
    if head.startswith(b"%PDF") or ext == ".pdf":
        t, f_, facts = pdf(path, rel, runrec)
        traits |= t
        features |= f_
    elif ext in (".wav", ".mp3", ".m4a", ".m4b", ".ogg", ".opus", ".flac", ".mp4", ".mkv", ".webm", ".avi", ".mpeg", ".ogv"):
        t, f_, facts = audio_video(path)
        traits |= t
        features |= f_
    elif entry["format"].startswith("image/") or ext in (".emf", ".wmf", ".tif", ".heic", ".avif", ".webp", ".svg"):
        if ext != ".svg":
            t, f_, facts = image(path)
            traits |= t
            features |= f_
    elif ext == ".eml":
        with open(path, "rb") as f:
            t, f_, text, facts = message(f.read())
        traits |= t
        features |= f_
    elif ext == ".mbox":
        t, f_, text, facts = mbox(path)
        traits |= t
        features |= f_
    elif ext in (".msg", ".pst"):
        t, f_, facts = mail_records(runrec.records(rel, "mail"))
        traits |= t
        features |= f_
        if ext == ".msg":
            facts["body_properties"] = msg_bodies(path)
            if facts["body_properties"] == ["1013"]:
                features.add("feature:html-body")
    elif ext in (".one", ".onetoc2"):
        outcomes = {e.get("outcome") for e in runrec.events.get(rel, [])}
        t, f_, facts = onenote(runrec.records(rel, "onenote"), outcomes)
        traits |= t
        features |= f_
    elif ext in (".odt", ".ods", ".odp", ".odg"):
        text = odf_text(path)
    elif ext in (".zip", ".7z", ".gz", ".zst"):
        entries = archive_entries(path)
        facts = {"entries": len(entries)}
        if len(entries) >= 1000:
            traits.add("trait:scale.many-entries")
        if library & {"kind:zip", "kind:7z", "kind:gzip", "kind:zstd", "kind:tar"} and any(
            re.search(r"\.(zip|7z|tar|gz|tgz|zst|rar)$", e, re.I) for e in entries
        ):
            features.add("feature:nested-archives")
    if ext[1:] in OOXML_KINDS or ext in LEGACY_TARGET:
        target = path if ext[1:] in OOXML_KINDS else converted(path, tmp) if ext in LEGACY_TARGET else None
        if target:
            f_, ofacts = ooxml(target)
            features |= f_
            facts.update(ofacts, converted=target != path)
    if ext in (".txt", ".csv", ".tsv", ".md", ".json", ".ndjson", ".xml", ".srt", ".html"):
        with open(path, "rb") as f:
            data = f.read()
        trait, guess = charset(data)
        if trait:
            traits.add(trait)
        facts["charset"] = guess
        text = decode_text(data, guess.lower() if guess not in ("bom", "utf-8") else guess)
        decl = re.search(rb'<\?xml[^>]*encoding=["\']([\w.-]+)', data[:200])
        if decl and decl.group(1).decode().lower() not in UTF8:
            traits.add("trait:charset.declared")
    if text is None:
        text = runrec.text(rel)
    elif ext in (".eml", ".mbox"):
        text = text + "\n" + runrec.text(rel)
    s, counts = scripts_in(text)
    traits |= s
    facts["scripts"] = {k: v for k, v in counts.items() if v >= 5}
    return traits, features, facts


def main():
    with open(os.path.join(ROOT, "docs/coverage.json")) as f:
        coverage = json.load(f)
    with open(os.path.join(DOCS, "MANIFEST.json")) as f:
        manifest = {e["path"]: e for e in json.load(f)["files"]}
    runrec = Run(coverage["library_placement"]["run"])
    tmp = tempfile.mkdtemp()
    out = []
    try:
        for sample in sorted(coverage["samples"], key=lambda s: s["path"]):
            traits, features, facts = probe(sample, manifest[sample["path"]], runrec, tmp)
            placed = traits | features
            placed |= set(sample.get("judged", {}))
            row = {"path": sample["path"],
                   "traits": sorted(i for i in placed if i.startswith("trait:")),
                   "features": sorted(i for i in placed if i.startswith("feature:"))}
            if "--facts" in sys.argv:
                row["facts"] = facts
                row["probed"] = sorted(traits | features)
            out.append(row)
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    json.dump(out, sys.stdout, indent=2, ensure_ascii=False)
    sys.stdout.write("\n")


if __name__ == "__main__":
    main()
