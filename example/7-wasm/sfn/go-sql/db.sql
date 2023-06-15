-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS message_id_seq;

-- Table Definition
CREATE TABLE "public"."message" (
    "id" int4 NOT NULL DEFAULT nextval('message_id_seq'::regclass),
    "msg" text,
    "created_at" timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY ("id")
);
