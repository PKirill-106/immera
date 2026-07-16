# Immera Product Context

## Overview

Immera is a mobile e-reading application for people who read books and documents in foreign languages.

The application reduces interruptions caused by switching between a reader, translator, dictionary, and notes application.

The primary client will be a React Native mobile application. This repository contains the Go backend.

## Target users

The initial target audience includes:

- language learners approximately between A2 and C1 levels;
- readers of foreign-language fiction;
- readers of professional and educational materials;
- users who want to build vocabulary while reading in context.

## Core user problem

When users encounter an unfamiliar word, they often leave the reading application, open an external translator, copy the word or sentence, inspect the result, and then manually save it somewhere.

This interrupts reading flow and loses the context in which the word appeared.

Immera should make contextual translation and vocabulary collection part of the reading experience.

## Core product capabilities

### Document reading

Users should be able to:

- upload supported books or documents;
- open and read their documents;
- resume reading from the latest position;
- view document metadata and cover;
- track reading progress.

### Contextual translation

Users should be able to:

- select or tap a word in a sentence;
- receive a translation appropriate to the sentence context;
- receive a translation of the complete sentence;
- avoid repeated translation costs through a shared translation cache.

### Personal dictionary

Users should be able to:

- save translated words;
- preserve the original word form;
- preserve the original sentence;
- preserve the translated word and sentence;
- associate a saved word with its source document;
- organize dictionary entries into groups.

### Reading notes

The product may later support:

- notes attached to a paragraph or reading position;
- bookmarks;
- highlights;
- personal thoughts attached to document fragments.

## Initial product scope

The first backend stages will focus on:

1. application foundation;
2. authentication;
3. document management;
4. reading progress;
5. contextual translation;
6. personal dictionary.

The first implementation should prioritize correctness and maintainability over feature breadth.

## Out of scope for the initial backend

The following features are not part of the initial implementation unless explicitly requested:

- social functionality;
- public book marketplace;
- collaborative notes;
- recommendation engine;
- complex spaced-repetition algorithms;
- real-time collaboration;
- independent microservices for every business module.

## External translation service

Context-aware translation will eventually be performed by a separate Python FastAPI service.

The Python service will:

- receive a word and sentence context;
- call an LLM or another translation provider;
- use structured outputs;
- return normalized translation data.

The Go backend remains responsible for:

- authentication and authorization;
- request validation;
- cache-key generation;
- persistence;
- orchestration;
- user dictionary operations;
- exposing the public API to the mobile client.

## Product principles

Immera should optimize for:

1. uninterrupted reading flow;
2. context-preserving vocabulary collection;
3. fast repeated translations;
4. reliable user data persistence;
5. simple and accessible mobile UX.