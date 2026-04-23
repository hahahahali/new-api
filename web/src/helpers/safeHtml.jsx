/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useMemo } from 'react';
import { marked } from 'marked';

const SAFE_TAGS = new Set([
  'a',
  'abbr',
  'b',
  'blockquote',
  'br',
  'code',
  'del',
  'details',
  'div',
  'em',
  'figcaption',
  'figure',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'hr',
  'i',
  'img',
  'kbd',
  'li',
  'ol',
  'p',
  'pre',
  's',
  'section',
  'small',
  'span',
  'strong',
  'sub',
  'summary',
  'sup',
  'table',
  'tbody',
  'td',
  'th',
  'thead',
  'tr',
  'u',
  'ul',
]);

const DROP_TAGS = new Set([
  'base',
  'button',
  'embed',
  'form',
  'frame',
  'frameset',
  'head',
  'iframe',
  'input',
  'link',
  'meta',
  'object',
  'script',
  'select',
  'style',
  'svg',
  'textarea',
  'title',
]);

const TAG_ATTRS = {
  a: new Set(['href', 'rel', 'target', 'title']),
  details: new Set(['open']),
  img: new Set(['alt', 'height', 'loading', 'src', 'title', 'width']),
  ol: new Set(['start']),
  td: new Set(['colspan', 'rowspan']),
  th: new Set(['colspan', 'rowspan']),
};

const GLOBAL_ATTRS = new Set(['class', 'dir', 'lang', 'role', 'title']);
const SAFE_PROTOCOLS = new Set(['http:', 'https:', 'mailto:', 'tel:']);

const isRelativeUrl = (value) => {
  return (
    value.startsWith('/') ||
    value.startsWith('./') ||
    value.startsWith('../') ||
    value.startsWith('#') ||
    value.startsWith('?')
  );
};

const isSafeUrl = (value) => {
  const trimmedValue = typeof value === 'string' ? value.trim() : '';
  if (!trimmedValue) {
    return false;
  }
  if (trimmedValue.startsWith('//')) {
    return false;
  }
  if (isRelativeUrl(trimmedValue)) {
    return true;
  }

  try {
    const parsed = new URL(trimmedValue, window.location.origin);
    return SAFE_PROTOCOLS.has(parsed.protocol);
  } catch {
    return false;
  }
};

const unwrapElement = (element) => {
  const parent = element.parentNode;
  if (!parent) {
    return;
  }
  while (element.firstChild) {
    parent.insertBefore(element.firstChild, element);
  }
  parent.removeChild(element);
};

const sanitizeAttributes = (element, tagName) => {
  const allowedAttrs = new Set([
    ...GLOBAL_ATTRS,
    ...(TAG_ATTRS[tagName] ? Array.from(TAG_ATTRS[tagName]) : []),
  ]);

  Array.from(element.attributes).forEach((attr) => {
    const name = attr.name.toLowerCase();
    const value = attr.value;
    const isAllowedDataAttr = name.startsWith('data-');
    const isAllowedAriaAttr = name.startsWith('aria-');

    if (
      name.startsWith('on') ||
      name === 'style' ||
      (!allowedAttrs.has(name) && !isAllowedDataAttr && !isAllowedAriaAttr)
    ) {
      element.removeAttribute(attr.name);
      return;
    }

    if ((name === 'href' || name === 'src') && !isSafeUrl(value)) {
      element.removeAttribute(attr.name);
    }
  });

  if (tagName === 'a') {
    if (!element.getAttribute('href')) {
      element.removeAttribute('target');
      element.removeAttribute('rel');
      return;
    }
    const target = element.getAttribute('target');
    if (target && target !== '_blank' && target !== '_self') {
      element.removeAttribute('target');
    }
    element.setAttribute('rel', 'noopener noreferrer');
  }

  if (tagName === 'img') {
    if (!element.getAttribute('src')) {
      element.remove();
      return;
    }
    element.setAttribute('loading', 'lazy');
    element.setAttribute('referrerpolicy', 'no-referrer');
  }
};

const sanitizeTree = (root) => {
  Array.from(root.childNodes).forEach((node) => {
    if (node.nodeType === 8) {
      node.remove();
      return;
    }

    if (node.nodeType !== 1) {
      return;
    }

    const element = node;
    const tagName = element.tagName.toLowerCase();

    if (DROP_TAGS.has(tagName)) {
      element.remove();
      return;
    }

    sanitizeTree(element);

    if (!SAFE_TAGS.has(tagName)) {
      unwrapElement(element);
      return;
    }

    sanitizeAttributes(element, tagName);
  });
};

export const sanitizeRichTextHtml = (html) => {
  if (!html || typeof html !== 'string') {
    return '';
  }
  if (typeof document === 'undefined') {
    return html;
  }

  const template = document.createElement('template');
  template.innerHTML = html;
  sanitizeTree(template.content);
  return template.innerHTML;
};

export const renderMarkdownToSafeHtml = (markdown) => {
  if (!markdown || typeof markdown !== 'string') {
    return '';
  }
  return sanitizeRichTextHtml(marked.parse(markdown));
};

export const SafeHtml = ({ as: Component = 'div', html = '', ...props }) => {
  const sanitizedHtml = useMemo(() => sanitizeRichTextHtml(html), [html]);
  return (
    <Component {...props} dangerouslySetInnerHTML={{ __html: sanitizedHtml }} />
  );
};
