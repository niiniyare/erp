--
-- PostgreSQL database dump
--

-- Dumped from database version 17.0
-- Dumped by pg_dump version 17.0

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: fuzzystrmatch; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS fuzzystrmatch WITH SCHEMA public;


--
-- Name: entity_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.entity_type AS ENUM (
    'COMPANY',
    'REGIONAL',
    'DEPARTMENT',
    'COST_CENTER'
);


--
-- Name: financial_year_status_enum; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.financial_year_status_enum AS ENUM (
    'OPEN',
    'CURRENT_FINANCIAL_YEAR',
    'CLOSED'
);


--
-- Name: avgcost(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.avgcost(integer) RETURNS double precision
    LANGUAGE plpgsql
    AS $_$

DECLARE

v_cost float;
v_qty float;
v_parts_id alias for $1;

BEGIN

  SELECT INTO v_cost, v_qty SUM(i.sellprice * i.qty), SUM(i.qty)
  FROM invoice i
  JOIN ap a ON (a.id = i.trans_id)
  WHERE i.parts_id = v_parts_id;

  IF v_cost IS NULL THEN
    v_cost := 0;
  END IF;

  IF NOT v_qty IS NULL THEN
    IF v_qty = 0 THEN
      v_cost := 0;
    ELSE
      v_cost := v_cost/v_qty;
    END IF;
  END IF;

RETURN v_cost;
END;
$_$;


--
-- Name: check_department(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_department() RETURNS trigger
    LANGUAGE plpgsql
    AS $$

declare
  dpt_id int;

begin
 
  if new.department_id = 0 then
    delete from dpt_trans where trans_id = new.id;
    return NULL;
  end if;

  select into dpt_id trans_id from dpt_trans where trans_id = new.id;
  
  if dpt_id > 0 then
    update dpt_trans set department_id = new.department_id where trans_id = dpt_id;
  else
    insert into dpt_trans (trans_id, department_id) values (new.id, new.department_id);
  end if;
return NULL;

end;
$$;


--
-- Name: del_customer(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.del_customer() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from shipto where trans_id = old.id;
  delete from customertax where customer_id = old.id;
  delete from partscustomer where customer_id = old.id;
  delete from address where trans_id = old.id;
  return NULL;
end;
$$;


--
-- Name: del_recurring(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.del_recurring() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from recurring where id = old.id;
  delete from recurringemail where id = old.id;
  delete from recurringprint where id = old.id;
  return NULL;
end;
$$;


--
-- Name: del_vendor(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.del_vendor() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from shipto where trans_id = old.id;
  delete from vendortax where vendor_id = old.id;
  delete from partsvendor where vendor_id = old.id;
  delete from address where trans_id = old.id;
  return NULL;
end;
$$;


--
-- Name: del_yearend(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.del_yearend() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
begin
  delete from yearend where trans_id = old.id;
  return NULL;
end;
$$;


--
-- Name: lastcost(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.lastcost(integer) RETURNS double precision
    LANGUAGE plpgsql
    AS $_$
 
DECLARE 
  
v_cost float; 
v_parts_id alias for $1; 
   
BEGIN 
    
  SELECT INTO v_cost sellprice FROM invoice i
  JOIN ap a ON (a.id = i.trans_id) 
  WHERE i.parts_id = v_parts_id 
  ORDER BY a.transdate desc, a.id desc 
  LIMIT 1; 
 
  IF v_cost IS NULL THEN 
    v_cost := 0; 
  END IF; 

RETURN v_cost; 
END; 
$_$;


--
-- Name: to_filtered_tsvector(text, text, regconfig); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.to_filtered_tsvector(input text, filter_type text DEFAULT 'COMMON'::text, ts_config regconfig DEFAULT 'simple'::regconfig) RETURNS tsvector
    LANGUAGE plpgsql
    AS $$ DECLARE filtered_input TEXT; BEGIN SELECT string_agg(input_word, ' ') INTO filtered_input FROM unnest(string_to_array(input, ' ')) AS input_word WHERE lower(input_word) NOT IN (SELECT lower(word) FROM search_irrelevant_words WHERE search_target = 'COMMON' OR search_target = filter_type); RETURN to_tsvector(ts_config, filtered_input); END; $$;


--
-- Name: update_entity_paths(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_entity_paths() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    -- Self-reference
    INSERT INTO hierarchy_paths (ancestor_id, descendant_id, depth)
    VALUES (NEW.id, NEW.id, 0);
    
    -- Parent paths
    IF NEW.parent_id IS NOT NULL THEN
        INSERT INTO hierarchy_paths (ancestor_id, descendant_id, depth)
        SELECT p.ancestor_id, NEW.id, p.depth + 1
        FROM hierarchy_paths p
        WHERE p.descendant_id = NEW.parent_id;
    END IF;
    
    RETURN NEW;
END;
$$;


--
-- Name: update_timestamps(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_timestamps() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


--
-- Name: entry_id; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.entry_id
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: acc_trans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.acc_trans (
    trans_id integer,
    chart_id integer NOT NULL,
    amount double precision,
    transdate date DEFAULT CURRENT_DATE,
    source text,
    approved boolean DEFAULT true,
    fx_transaction boolean DEFAULT false,
    project_id integer,
    memo text,
    id integer,
    cleared boolean DEFAULT false,
    vr_id integer,
    entry_id integer DEFAULT nextval('public.entry_id'::regclass),
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date
);


--
-- Name: acc_trans_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.acc_trans_log (
    trans_id integer,
    chart_id integer,
    amount double precision,
    transdate date,
    source text,
    approved boolean,
    fx_transaction boolean,
    project_id integer,
    memo text,
    id integer,
    cleared boolean,
    vr_id integer,
    entry_id integer,
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date,
    ts timestamp without time zone DEFAULT now()
);


--
-- Name: acc_trans_log_deleted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.acc_trans_log_deleted (
    trans_id integer,
    chart_id integer,
    amount double precision,
    transdate date,
    source text,
    approved boolean,
    fx_transaction boolean,
    project_id integer,
    memo text,
    id integer,
    cleared boolean,
    vr_id integer,
    entry_id integer,
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date,
    ts timestamp without time zone DEFAULT now()
);


--
-- Name: addressid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.addressid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: address; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.address (
    id integer DEFAULT nextval('public.addressid'::regclass) NOT NULL,
    trans_id integer,
    address1 character varying(64),
    address2 character varying(64),
    city character varying(64),
    is_migrated boolean DEFAULT true,
    post_office character varying(64),
    state character varying(32),
    zipcode character varying(32),
    country character varying(32)
);


--
-- Name: id; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.id
    START WITH 10000
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ap; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ap (
    id integer DEFAULT nextval('public.id'::regclass),
    invnumber text,
    transdate date DEFAULT CURRENT_DATE,
    vendor_id integer,
    taxincluded boolean DEFAULT false,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean DEFAULT false,
    ordnumber text,
    curr character(3),
    notes text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer DEFAULT 0,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    shippingpoint text,
    terms smallint DEFAULT 0,
    approved boolean DEFAULT true,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean DEFAULT false,
    ts timestamp without time zone DEFAULT now(),
    exchangerate double precision,
    dcn text
);


--
-- Name: ap_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ap_log (
    id integer,
    invnumber text,
    transdate date,
    vendor_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    ordnumber text,
    curr character(3),
    notes text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    shippingpoint text,
    terms smallint,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);


--
-- Name: ap_log_deleted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ap_log_deleted (
    id integer,
    invnumber text,
    transdate date,
    vendor_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    ordnumber text,
    curr character(3),
    notes text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    shippingpoint text,
    terms smallint,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);


--
-- Name: ar; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar (
    id integer DEFAULT nextval('public.id'::regclass),
    invnumber text,
    transdate date DEFAULT CURRENT_DATE,
    customer_id integer,
    taxincluded boolean DEFAULT false,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean DEFAULT false,
    shippingpoint text,
    terms smallint,
    notes text,
    curr character(3),
    ordnumber text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer DEFAULT 0,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    approved boolean DEFAULT true,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean DEFAULT false,
    ts timestamp without time zone DEFAULT now(),
    exchangerate double precision,
    dcn text
);


--
-- Name: ar_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar_log (
    id integer,
    invnumber text,
    transdate date,
    customer_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    shippingpoint text,
    terms smallint,
    notes text,
    curr character(3),
    ordnumber text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);


--
-- Name: ar_log_deleted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ar_log_deleted (
    id integer,
    invnumber text,
    transdate date,
    customer_id integer,
    taxincluded boolean,
    amount double precision,
    netamount double precision,
    paid double precision,
    datepaid date,
    duedate date,
    invoice boolean,
    shippingpoint text,
    terms smallint,
    notes text,
    curr character(3),
    ordnumber text,
    till character varying(20),
    quonumber text,
    intnotes text,
    department_id integer,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    approved boolean,
    cashdiscount real,
    discountterms smallint,
    waybill text,
    warehouse_id integer,
    linetax boolean,
    ts timestamp without time zone
);


--
-- Name: assemblyid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.assemblyid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: assembly; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.assembly (
    id integer DEFAULT nextval('public.assemblyid'::regclass),
    parts_id integer,
    qty double precision,
    bom boolean,
    adj boolean,
    aid integer
);


--
-- Name: audittrail; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audittrail (
    trans_id integer,
    tablename text,
    reference text,
    formname text,
    action text,
    transdate timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    employee_id integer
);


--
-- Name: bank; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bank (
    id integer,
    name character varying(64),
    iban character varying(34),
    bic character varying(11),
    address_id integer DEFAULT nextval('public.addressid'::regclass),
    dcn text,
    rvc text,
    strdbkginf text,
    invdescriptionqr text,
    qriban text,
    membernumber text
);


--
-- Name: bank_account; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.bank_account (
    id integer NOT NULL,
    bic text,
    iban text,
    account text,
    bank text,
    branch text,
    country text,
    country_iso text,
    city text,
    state text,
    zip text,
    address text,
    description text,
    chart_id integer,
    inactive boolean,
    blink_permission_id bigint,
    blink_account_id text
);


--
-- Name: bank_account_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.bank_account ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.bank_account_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: banking_import_event; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.banking_import_event (
    id bigint NOT NULL,
    delegated_to text
);


--
-- Name: banking_import_event_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.banking_import_event ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.banking_import_event_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: blink_import_process; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blink_import_process (
    id bigint NOT NULL,
    bank_account_id bigint NOT NULL,
    status text NOT NULL,
    error jsonb,
    created_at timestamp without time zone NOT NULL,
    last_modified_at timestamp without time zone
);


--
-- Name: blink_import_process_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.blink_import_process ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.blink_import_process_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: blink_import_process_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.blink_import_process_log (
    id bigint NOT NULL,
    blink_import_process_id bigint NOT NULL,
    processed_target_id text NOT NULL,
    processed_payload jsonb NOT NULL,
    error jsonb,
    banking_import_event_id bigint
);


--
-- Name: blink_import_process_log_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.blink_import_process_log ALTER COLUMN id ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.blink_import_process_log_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: booking_to_settlement; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.booking_to_settlement (
    id integer NOT NULL,
    booking_id integer NOT NULL,
    settlement_id integer NOT NULL
);


--
-- Name: booking_to_settlement_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.booking_to_settlement_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: booking_to_settlement_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.booking_to_settlement_id_seq OWNED BY public.booking_to_settlement.id;


--
-- Name: br; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.br (
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    batchnumber text,
    description text,
    batch text,
    transdate date DEFAULT CURRENT_DATE,
    apprdate date,
    amount double precision,
    managerid integer,
    employee_id integer
);


--
-- Name: build; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.build (
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    reference text,
    transdate date,
    department_id integer,
    warehouse_id integer,
    employee_id integer
);


--
-- Name: business; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.business (
    id integer DEFAULT nextval('public.id'::regclass),
    description text,
    discount real
);


--
-- Name: cargo; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cargo (
    id integer NOT NULL,
    trans_id integer NOT NULL,
    package text,
    netweight double precision,
    grossweight double precision,
    volume double precision
);


--
-- Name: chart; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chart (
    id integer DEFAULT nextval('public.id'::regclass),
    accno integer,
    description text,
    balance double precision,
    type character(1),
    gifi integer,
    category character(1),
    link text,
    gifi_accno text,
    contra boolean DEFAULT false,
    parent_id integer,
    charttype character(1) DEFAULT 'A'::bpchar
);


--
-- Name: chat; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.chat (
    id integer NOT NULL,
    trans_id integer NOT NULL,
    message character varying(255) NOT NULL,
    employee_id integer NOT NULL,
    creation_date timestamp without time zone NOT NULL
);


--
-- Name: chat_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.chat_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: chat_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.chat_id_seq OWNED BY public.chat.id;


--
-- Name: contactid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contactid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contact; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contact (
    id integer DEFAULT nextval('public.contactid'::regclass) NOT NULL,
    trans_id integer NOT NULL,
    salutation character varying(32),
    firstname character varying(32),
    lastname character varying(32),
    contacttitle character varying(32),
    occupation character varying(32),
    phone character varying(20),
    fax character varying(20),
    mobile character varying(20),
    email text,
    gender character(1) DEFAULT 'M'::bpchar,
    parent_id integer,
    typeofcontact character varying(20)
);


--
-- Name: credits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.credits (
    id integer NOT NULL,
    reference text,
    description text,
    transdate date,
    accno text,
    amount numeric(12,2)
);


--
-- Name: credits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.credits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: credits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.credits_id_seq OWNED BY public.credits.id;


--
-- Name: curr; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.curr (
    rn integer,
    curr character(3) NOT NULL,
    "precision" smallint
);


--
-- Name: customer; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customer (
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    name character varying(64),
    contact character varying(64),
    phone character varying(20),
    fax character varying(20),
    email text,
    notes text,
    terms smallint DEFAULT 0,
    taxincluded boolean DEFAULT false,
    customernumber character varying(32),
    cc text,
    bcc text,
    business_id integer,
    taxnumber character varying(32),
    sic_code character varying(6),
    discount real,
    creditlimit double precision DEFAULT 0,
    iban character varying(34),
    bic character varying(11),
    employee_id integer,
    language_code character varying(6),
    pricegroup_id integer,
    curr character(3),
    startdate date,
    enddate date,
    arap_accno_id integer,
    payment_accno_id integer,
    discount_accno_id integer,
    cashdiscount real,
    discountterms smallint,
    threshold double precision,
    dispatch_id integer
);


--
-- Name: customercart; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customercart (
    cart_id character varying(32),
    customer_id integer,
    parts_id integer,
    qty double precision DEFAULT 0,
    price double precision DEFAULT 0,
    taxaccounts text
);


--
-- Name: customerloginid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.customerloginid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: customerlogin; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customerlogin (
    id integer DEFAULT nextval('public.customerloginid'::regclass) NOT NULL,
    login character varying(100),
    passwd character(32),
    session character(32),
    session_exp character varying(20),
    customer_id integer
);


--
-- Name: customertax; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customertax (
    customer_id integer,
    chart_id integer
);


--
-- Name: debits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.debits (
    id integer NOT NULL,
    reference text,
    description text,
    transdate date,
    accno text,
    amount numeric(12,2)
);


--
-- Name: debits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.debits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: debits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.debits_id_seq OWNED BY public.debits.id;


--
-- Name: debitscredits; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.debitscredits (
    id integer NOT NULL,
    reference text,
    description text,
    transdate date,
    debit_accno text,
    credit_accno text,
    amount numeric(12,2)
);


--
-- Name: debitscredits_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.debitscredits_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: debitscredits_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.debitscredits_id_seq OWNED BY public.debitscredits.id;


--
-- Name: defaults; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.defaults (
    fldname text,
    fldvalue text
);


--
-- Name: department; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.department (
    id integer DEFAULT nextval('public.id'::regclass),
    description text,
    role character(1) DEFAULT 'P'::bpchar
);


--
-- Name: dispatch; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dispatch (
    id integer DEFAULT nextval('public.id'::regclass),
    description text
);


--
-- Name: dpt_trans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dpt_trans (
    trans_id integer,
    department_id integer
);


--
-- Name: employee; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.employee (
    id integer DEFAULT nextval('public.id'::regclass),
    login text,
    name character varying(64),
    address1 character varying(32),
    address2 character varying(32),
    city character varying(32),
    state character varying(32),
    zipcode character varying(10),
    country character varying(32),
    workphone character varying(20),
    homephone character varying(20),
    startdate date DEFAULT CURRENT_DATE,
    enddate date,
    notes text,
    role character varying(20),
    sales boolean DEFAULT false,
    email text,
    ssn character varying(20),
    iban character varying(34),
    bic character varying(11),
    managerid integer,
    employeenumber character varying(32),
    dob date
);


--
-- Name: entities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entities (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    type public.entity_type NOT NULL,
    parent_id integer,
    organization_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: entities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.entities_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: entities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.entities_id_seq OWNED BY public.entities.id;


--
-- Name: exchangerate; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.exchangerate (
    curr character(3),
    transdate date,
    buy double precision,
    sell double precision
);


--
-- Name: fifo; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fifo (
    trans_id integer,
    transdate date,
    parts_id integer,
    qty double precision,
    costprice double precision,
    sellprice double precision,
    warehouse_id integer,
    invoice_id integer
);


--
-- Name: filtered; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.filtered (
    trans_id integer,
    entry_id integer DEFAULT 0,
    entry_id2 integer DEFAULT 0,
    debit double precision DEFAULT 0,
    credit double precision DEFAULT 0
);


--
-- Name: financial_year; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.financial_year (
    id integer NOT NULL,
    start_date date,
    end_date date,
    status public.financial_year_status_enum,
    yearend_id integer
);


--
-- Name: financial_year_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.financial_year_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: financial_year_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.financial_year_id_seq OWNED BY public.financial_year.id;


--
-- Name: gifi; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gifi (
    accno text,
    description text
);


--
-- Name: gl; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gl (
    id integer DEFAULT nextval('public.id'::regclass),
    reference text,
    description text,
    transdate date DEFAULT CURRENT_DATE,
    employee_id integer,
    notes text,
    department_id integer DEFAULT 0,
    approved boolean DEFAULT true,
    curr character(3),
    exchangerate double precision,
    ts timestamp without time zone DEFAULT now(),
    onhold boolean
);


--
-- Name: gl_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gl_log (
    id integer,
    reference text,
    description text,
    transdate date,
    employee_id integer,
    notes text,
    department_id integer,
    approved boolean,
    curr character(3),
    exchangerate double precision,
    ts timestamp without time zone,
    onhold boolean
);


--
-- Name: gl_log_deleted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.gl_log_deleted (
    id integer,
    reference text,
    description text,
    transdate date,
    employee_id integer,
    notes text,
    department_id integer,
    approved boolean,
    curr character(3),
    exchangerate double precision,
    ts timestamp without time zone,
    onhold boolean
);


--
-- Name: hierarchy_paths; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.hierarchy_paths (
    ancestor_id integer NOT NULL,
    descendant_id integer NOT NULL,
    depth integer NOT NULL
);


--
-- Name: inventoryid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.inventoryid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: inventory; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.inventory (
    id integer DEFAULT nextval('public.inventoryid'::regclass),
    warehouse_id integer,
    parts_id integer,
    trans_id integer,
    orderitems_id integer,
    qty double precision,
    shippingdate date,
    employee_id integer,
    department_id integer,
    warehouse_id2 integer,
    serialnumber text,
    itemnotes text,
    cost double precision,
    linetype character(1) DEFAULT '0'::bpchar,
    description text,
    invoice_id integer,
    cogs double precision
);


--
-- Name: invoiceid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.invoiceid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: invoice; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice (
    id integer DEFAULT nextval('public.invoiceid'::regclass),
    trans_id integer,
    parts_id integer,
    description text,
    qty real,
    allocated real,
    sellprice double precision,
    fxsellprice double precision,
    discount real,
    assemblyitem boolean DEFAULT false,
    project_id integer,
    deliverydate date,
    serialnumber text,
    notes text
);


--
-- Name: invoice_log; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_log (
    id integer,
    trans_id integer,
    parts_id integer,
    description text,
    qty real,
    allocated real,
    sellprice double precision,
    fxsellprice double precision,
    discount real,
    assemblyitem boolean,
    project_id integer,
    deliverydate date,
    serialnumber text,
    notes text,
    ts timestamp without time zone DEFAULT now()
);


--
-- Name: invoice_log_deleted; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoice_log_deleted (
    trans_id integer,
    chart_id integer,
    amount double precision,
    transdate date,
    source text,
    approved boolean,
    fx_transaction boolean,
    project_id integer,
    memo text,
    id integer,
    cleared boolean,
    vr_id integer,
    entry_id integer,
    tax text,
    taxamount double precision,
    tax_chart_id integer,
    invoice_id integer,
    reconciled date,
    ts timestamp without time zone DEFAULT now()
);


--
-- Name: invoicetax; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.invoicetax (
    trans_id integer,
    invoice_id integer,
    chart_id integer,
    taxamount double precision,
    amount double precision
);


--
-- Name: jcitemsid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jcitemsid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jcitems; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jcitems (
    id integer DEFAULT nextval('public.jcitemsid'::regclass),
    project_id integer,
    parts_id integer,
    description text,
    qty double precision,
    allocated double precision,
    sellprice double precision,
    fxsellprice double precision,
    serialnumber text,
    checkedin timestamp with time zone,
    checkedout timestamp with time zone,
    employee_id integer,
    notes text
);


--
-- Name: language; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.language (
    code character varying(6),
    description text
);


--
-- Name: lastused; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.lastused (
    id integer NOT NULL,
    report character varying(40),
    cols text,
    login character varying(255)
);


--
-- Name: lastused_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.lastused_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: lastused_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.lastused_id_seq OWNED BY public.lastused.id;


--
-- Name: makemodel; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.makemodel (
    parts_id integer,
    make text,
    model text
);


--
-- Name: oe; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oe (
    id integer DEFAULT nextval('public.id'::regclass),
    ordnumber text,
    transdate date DEFAULT CURRENT_DATE,
    vendor_id integer,
    customer_id integer,
    amount double precision,
    netamount double precision,
    reqdate date,
    taxincluded boolean,
    shippingpoint text,
    notes text,
    curr character(3),
    employee_id integer,
    closed boolean DEFAULT false,
    quotation boolean DEFAULT false,
    quonumber text,
    intnotes text,
    department_id integer DEFAULT 0,
    shipvia text,
    language_code character varying(6),
    ponumber text,
    terms smallint DEFAULT 0,
    waybill text,
    warehouse_id integer,
    description text,
    aa_id integer,
    exchangerate double precision
);


--
-- Name: oldchart; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.oldchart (
    id integer DEFAULT nextval('public.id'::regclass),
    accno text NOT NULL,
    description text,
    parent_id integer,
    charttype character(1) DEFAULT 'A'::bpchar,
    category character(1),
    link text,
    gifi_accno text,
    contra boolean DEFAULT false,
    allow_gl boolean DEFAULT true,
    symbol_link character varying(128)
);


--
-- Name: orderitemsid; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orderitemsid
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: orderitems; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orderitems (
    id integer DEFAULT nextval('public.orderitemsid'::regclass),
    trans_id integer,
    parts_id integer,
    description text,
    qty double precision,
    sellprice double precision,
    discount real,
    unit character varying(5),
    project_id integer,
    reqdate date,
    ship double precision,
    serialnumber text,
    itemnotes text,
    lineitemdetail boolean,
    ordernumber text,
    ponumber text,
    notes text
);


--
-- Name: org; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.org (
    id integer NOT NULL,
    name character varying(255) NOT NULL,
    hierarchy_level integer DEFAULT 1 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: org_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.org_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: org_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.org_id_seq OWNED BY public.org.id;


--
-- Name: parts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.parts (
    id integer DEFAULT nextval('public.id'::regclass),
    partnumber text,
    description text,
    unit character varying(5),
    listprice double precision,
    sellprice double precision,
    lastcost double precision,
    priceupdate date DEFAULT CURRENT_DATE,
    weight double precision,
    onhand double precision DEFAULT 0,
    notes text,
    makemodel boolean DEFAULT false,
    assembly boolean DEFAULT false,
    alternate boolean DEFAULT false,
    rop double precision,
    inventory_accno_id integer,
    income_accno_id integer,
    expense_accno_id integer,
    bin text,
    obsolete boolean DEFAULT false,
    bom boolean DEFAULT false,
    image text,
    drawing text,
    microfiche text,
    partsgroup_id integer,
    project_id integer,
    avgcost double precision,
    tariff_hscode text,
    countryorigin text,
    barcode text,
    toolnumber text
);


--
-- Name: partsattr; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.partsattr (
    parts_id integer,
    hotnew character varying(3)
);


--
-- Name: partscustomer; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.partscustomer (
    parts_id integer,
    customer_id integer,
    pricegroup_id integer,
    pricebreak double precision,
    sellprice double precision,
    validfrom date,
    validto date,
    curr character(3)
);


--
-- Name: partsgroup; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.partsgroup (
    id integer DEFAULT nextval('public.id'::regclass),
    partsgroup text,
    pos boolean DEFAULT true
);


--
-- Name: partstax; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.partstax (
    parts_id integer,
    chart_id integer
);


--
-- Name: partsvendor; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.partsvendor (
    vendor_id integer,
    parts_id integer,
    partnumber text,
    leadtime smallint,
    lastcost double precision,
    curr character(3)
);


--
-- Name: payment; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payment (
    id integer NOT NULL,
    trans_id integer NOT NULL,
    exchangerate double precision DEFAULT 1,
    paymentmethod_id integer
);


--
-- Name: paymentmethod; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.paymentmethod (
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    description text,
    fee double precision,
    rn integer
);


--
-- Name: pricegroup; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pricegroup (
    id integer DEFAULT nextval('public.id'::regclass),
    pricegroup text
);


--
-- Name: project; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.project (
    id integer DEFAULT nextval('public.id'::regclass),
    projectnumber text,
    description text,
    startdate date,
    enddate date,
    parts_id integer,
    production double precision DEFAULT 0,
    completed double precision DEFAULT 0,
    customer_id integer
);


--
-- Name: recurring; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recurring (
    id integer,
    reference text,
    startdate date,
    nextdate date,
    enddate date,
    repeat smallint,
    unit character varying(6),
    howmany integer,
    payment boolean DEFAULT false,
    description text
);


--
-- Name: recurringemail; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recurringemail (
    id integer,
    formname text,
    format text,
    message text
);


--
-- Name: recurringprint; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recurringprint (
    id integer,
    formname text,
    format text,
    printer text
);


--
-- Name: report; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.report (
    reportid integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    reportcode text,
    reportdescription text,
    login text
);


--
-- Name: reportvars; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reportvars (
    reportid integer NOT NULL,
    reportvariable text,
    reportvalue text
);


--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


--
-- Name: search_irrelevant_words; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.search_irrelevant_words (
    id integer NOT NULL,
    search_target text NOT NULL,
    word text NOT NULL
);


--
-- Name: search_irrelevant_words_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.search_irrelevant_words_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: search_irrelevant_words_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.search_irrelevant_words_id_seq OWNED BY public.search_irrelevant_words.id;


--
-- Name: semaphore; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.semaphore (
    id integer,
    login text,
    module text,
    expires character varying(10)
);


--
-- Name: shipto; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.shipto (
    trans_id integer,
    shiptoname character varying(64),
    shiptoaddress1 character varying(32),
    shiptoaddress2 character varying(32),
    shiptocity character varying(32),
    shiptostate character varying(32),
    shiptozipcode character varying(10),
    shiptocountry character varying(32),
    shiptocontact character varying(64),
    shiptophone character varying(20),
    shiptofax character varying(20),
    shiptoemail text
);


--
-- Name: sic; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sic (
    code character varying(6),
    sictype character(1),
    description text
);


--
-- Name: status; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.status (
    trans_id integer,
    formname text,
    printed boolean DEFAULT false,
    emailed boolean DEFAULT false,
    spoolfile text
);


--
-- Name: tax; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tax (
    chart_id integer,
    rate double precision,
    taxnumber text,
    reversecharge_id integer,
    validto date,
    vatkey character varying(5),
    formdigit integer DEFAULT 0,
    validfrom date,
    id integer NOT NULL
);


--
-- Name: tax_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tax_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tax_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tax_id_seq OWNED BY public.tax.id;


--
-- Name: translation; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.translation (
    trans_id integer,
    language_code character varying(6),
    description text
);


--
-- Name: trf; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.trf (
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    transdate date,
    trfnumber text,
    description text,
    notes text,
    department_id integer,
    from_warehouse_id integer,
    to_warehouse_id integer DEFAULT 0,
    employee_id integer DEFAULT 0,
    delivereddate date
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email character varying(255) NOT NULL,
    full_name character varying(255) NOT NULL,
    entity_id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    last_login_at timestamp with time zone,
    deleted_at timestamp with time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: v_active_users; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_active_users AS
 SELECT u.id AS user_id,
    u.email,
    u.full_name,
    o.name AS organization,
    c.name AS company,
    r.name AS regional,
    d.name AS department,
    cc.name AS cost_center,
    u.last_login_at
   FROM (((((public.users u
     JOIN public.entities cc ON ((u.entity_id = cc.id)))
     JOIN public.entities d ON ((cc.parent_id = d.id)))
     JOIN public.entities r ON ((d.parent_id = r.id)))
     JOIN public.entities c ON ((r.parent_id = c.id)))
     JOIN public.org o ON ((c.organization_id = o.id)))
  WHERE ((u.deleted_at IS NULL) AND (u.last_login_at > (CURRENT_DATE - '90 days'::interval)) AND (cc.deleted_at IS NULL) AND (d.deleted_at IS NULL) AND (r.deleted_at IS NULL) AND (c.deleted_at IS NULL) AND (o.deleted_at IS NULL));


--
-- Name: v_company_subtree; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_company_subtree AS
 SELECT c.id AS company_id,
    c.name AS company,
    e.id AS entity_id,
    e.name AS entity_name,
    e.type AS entity_type,
    hp.depth AS levels_from_company
   FROM ((public.entities c
     JOIN public.hierarchy_paths hp ON ((c.id = hp.ancestor_id)))
     JOIN public.entities e ON ((hp.descendant_id = e.id)))
  WHERE ((c.type = 'COMPANY'::public.entity_type) AND (c.deleted_at IS NULL) AND (e.deleted_at IS NULL));


--
-- Name: v_cost_center_management; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_cost_center_management AS
 SELECT cc.id AS cost_center_id,
    cc.name AS cost_center,
    d.id AS department_id,
    d.name AS department,
    r.id AS regional_id,
    r.name AS regional,
    c.id AS company_id,
    c.name AS company,
    o.id AS organization_id,
    o.name AS organization,
    count(u.id) AS user_count,
    string_agg((u.full_name)::text, ', '::text) AS users
   FROM (((((public.entities cc
     JOIN public.entities d ON ((cc.parent_id = d.id)))
     JOIN public.entities r ON ((d.parent_id = r.id)))
     JOIN public.entities c ON ((r.parent_id = c.id)))
     JOIN public.org o ON ((c.organization_id = o.id)))
     LEFT JOIN public.users u ON ((u.entity_id = cc.id)))
  WHERE ((cc.type = 'COST_CENTER'::public.entity_type) AND (cc.deleted_at IS NULL))
  GROUP BY cc.id, cc.name, d.id, d.name, r.id, r.name, c.id, c.name, o.id, o.name;


--
-- Name: v_department_summary; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_department_summary AS
 SELECT d.id AS department_id,
    d.name AS department,
    r.id AS regional_id,
    r.name AS regional,
    c.id AS company_id,
    c.name AS company,
    o.id AS organization_id,
    o.name AS organization,
    count(DISTINCT cc.id) AS cost_center_count,
    count(DISTINCT u.id) AS user_count
   FROM (((((public.entities d
     JOIN public.entities r ON ((d.parent_id = r.id)))
     JOIN public.entities c ON ((r.parent_id = c.id)))
     JOIN public.org o ON ((c.organization_id = o.id)))
     LEFT JOIN public.entities cc ON (((cc.parent_id = d.id) AND (cc.type = 'COST_CENTER'::public.entity_type) AND (cc.deleted_at IS NULL))))
     LEFT JOIN public.users u ON (((u.entity_id = cc.id) AND (u.deleted_at IS NULL))))
  WHERE ((d.type = 'DEPARTMENT'::public.entity_type) AND (d.deleted_at IS NULL))
  GROUP BY d.id, d.name, r.id, r.name, c.id, c.name, o.id, o.name;


--
-- Name: v_organization_hierarchy; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_organization_hierarchy AS
 WITH RECURSIVE org_chart AS (
         SELECT entities.id,
            entities.name,
            entities.type,
            entities.parent_id,
            entities.organization_id,
            entities.deleted_at,
            (entities.name)::text AS path,
            0 AS depth
           FROM public.entities
          WHERE (entities.parent_id IS NULL)
        UNION ALL
         SELECT e.id,
            e.name,
            e.type,
            e.parent_id,
            e.organization_id,
            e.deleted_at,
            ((oc_1.path || ' > '::text) || (e.name)::text) AS text,
            (oc_1.depth + 1)
           FROM (public.entities e
             JOIN org_chart oc_1 ON ((e.parent_id = oc_1.id)))
        )
 SELECT o.name AS organization,
    oc.id AS entity_id,
    oc.name AS entity_name,
    oc.type AS entity_type,
    oc.path AS full_path,
    oc.depth
   FROM (org_chart oc
     JOIN public.org o ON ((oc.organization_id = o.id)))
  WHERE (oc.deleted_at IS NULL);


--
-- Name: v_organization_size_report; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_organization_size_report AS
 SELECT o.id AS organization_id,
    o.name AS organization,
    count(DISTINCT c.id) FILTER (WHERE ((c.type = 'COMPANY'::public.entity_type) AND (c.deleted_at IS NULL))) AS company_count,
    count(DISTINCT r.id) FILTER (WHERE ((r.type = 'REGIONAL'::public.entity_type) AND (r.deleted_at IS NULL))) AS regional_count,
    count(DISTINCT d.id) FILTER (WHERE ((d.type = 'DEPARTMENT'::public.entity_type) AND (d.deleted_at IS NULL))) AS department_count,
    count(DISTINCT cc.id) FILTER (WHERE ((cc.type = 'COST_CENTER'::public.entity_type) AND (cc.deleted_at IS NULL))) AS cost_center_count,
    count(DISTINCT u.id) FILTER (WHERE (u.deleted_at IS NULL)) AS user_count
   FROM (((((public.org o
     LEFT JOIN public.entities c ON ((c.organization_id = o.id)))
     LEFT JOIN public.entities r ON ((r.organization_id = o.id)))
     LEFT JOIN public.entities d ON ((d.organization_id = o.id)))
     LEFT JOIN public.entities cc ON ((cc.organization_id = o.id)))
     LEFT JOIN public.users u ON ((u.entity_id = cc.id)))
  WHERE (o.deleted_at IS NULL)
  GROUP BY o.id, o.name;


--
-- Name: v_user_directory; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.v_user_directory AS
 SELECT u.id AS user_id,
    u.email,
    u.full_name,
    u.last_login_at,
    cc.id AS cost_center_id,
    cc.name AS cost_center,
    d.id AS department_id,
    d.name AS department,
    r.id AS regional_id,
    r.name AS regional,
    c.id AS company_id,
    c.name AS company,
    o.id AS organization_id,
    o.name AS organization
   FROM (((((public.users u
     JOIN public.entities cc ON ((u.entity_id = cc.id)))
     JOIN public.entities d ON ((cc.parent_id = d.id)))
     JOIN public.entities r ON ((d.parent_id = r.id)))
     JOIN public.entities c ON ((r.parent_id = c.id)))
     JOIN public.org o ON ((c.organization_id = o.id)))
  WHERE (u.deleted_at IS NULL);


--
-- Name: vat_settlement; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vat_settlement (
    id integer NOT NULL,
    vat_form_id integer,
    period_from date,
    period_to date,
    creation_date timestamp without time zone,
    data json,
    xml bytea
);


--
-- Name: vat_settlement_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.vat_settlement_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: vat_settlement_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.vat_settlement_id_seq OWNED BY public.vat_settlement.id;


--
-- Name: vendor; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vendor (
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    name character varying(64),
    contact character varying(64),
    phone character varying(20),
    fax character varying(20),
    email text,
    notes text,
    terms smallint DEFAULT 0,
    taxincluded boolean DEFAULT false,
    vendornumber character varying(32),
    cc text,
    bcc text,
    gifi_accno character varying(30),
    business_id integer,
    taxnumber character varying(32),
    sic_code character varying(6),
    discount real,
    creditlimit double precision DEFAULT 0,
    iban character varying(34),
    bic character varying(11),
    employee_id integer,
    language_code character varying(6),
    pricegroup_id integer,
    curr character(3),
    startdate date,
    enddate date,
    arap_accno_id integer,
    payment_accno_id integer,
    discount_accno_id integer,
    cashdiscount real,
    discountterms smallint,
    threshold double precision,
    dispatch_id integer
);


--
-- Name: vendortax; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vendortax (
    vendor_id integer,
    chart_id integer
);


--
-- Name: vr; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vr (
    br_id integer,
    trans_id integer NOT NULL,
    id integer DEFAULT nextval('public.id'::regclass) NOT NULL,
    vouchernumber text
);


--
-- Name: warehouse; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.warehouse (
    id integer DEFAULT nextval('public.id'::regclass),
    description text
);


--
-- Name: yearend; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.yearend (
    trans_id integer,
    transdate date,
    id integer NOT NULL
);


--
-- Name: yearend_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.yearend_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: yearend_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.yearend_id_seq OWNED BY public.yearend.id;


--
-- Name: booking_to_settlement id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_to_settlement ALTER COLUMN id SET DEFAULT nextval('public.booking_to_settlement_id_seq'::regclass);


--
-- Name: chat id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat ALTER COLUMN id SET DEFAULT nextval('public.chat_id_seq'::regclass);


--
-- Name: credits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.credits ALTER COLUMN id SET DEFAULT nextval('public.credits_id_seq'::regclass);


--
-- Name: debits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.debits ALTER COLUMN id SET DEFAULT nextval('public.debits_id_seq'::regclass);


--
-- Name: debitscredits id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.debitscredits ALTER COLUMN id SET DEFAULT nextval('public.debitscredits_id_seq'::regclass);


--
-- Name: entities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entities ALTER COLUMN id SET DEFAULT nextval('public.entities_id_seq'::regclass);


--
-- Name: financial_year id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_year ALTER COLUMN id SET DEFAULT nextval('public.financial_year_id_seq'::regclass);


--
-- Name: lastused id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lastused ALTER COLUMN id SET DEFAULT nextval('public.lastused_id_seq'::regclass);


--
-- Name: org id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.org ALTER COLUMN id SET DEFAULT nextval('public.org_id_seq'::regclass);


--
-- Name: search_irrelevant_words id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_irrelevant_words ALTER COLUMN id SET DEFAULT nextval('public.search_irrelevant_words_id_seq'::regclass);


--
-- Name: tax id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tax ALTER COLUMN id SET DEFAULT nextval('public.tax_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: vat_settlement id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vat_settlement ALTER COLUMN id SET DEFAULT nextval('public.vat_settlement_id_seq'::regclass);


--
-- Name: yearend id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.yearend ALTER COLUMN id SET DEFAULT nextval('public.yearend_id_seq'::regclass);


--
-- Name: address address_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.address
    ADD CONSTRAINT address_pkey PRIMARY KEY (id);


--
-- Name: bank_account bank_account_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.bank_account
    ADD CONSTRAINT bank_account_pkey PRIMARY KEY (id);


--
-- Name: banking_import_event banking_import_event_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.banking_import_event
    ADD CONSTRAINT banking_import_event_pkey PRIMARY KEY (id);


--
-- Name: blink_import_process_log blink_import_process_log_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blink_import_process_log
    ADD CONSTRAINT blink_import_process_log_pkey PRIMARY KEY (id);


--
-- Name: blink_import_process blink_import_process_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blink_import_process
    ADD CONSTRAINT blink_import_process_pkey PRIMARY KEY (id);


--
-- Name: booking_to_settlement booking_to_settlement_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_to_settlement
    ADD CONSTRAINT booking_to_settlement_pkey PRIMARY KEY (id);


--
-- Name: br br_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.br
    ADD CONSTRAINT br_pkey PRIMARY KEY (id);


--
-- Name: build build_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.build
    ADD CONSTRAINT build_pkey PRIMARY KEY (id);


--
-- Name: chart chart_accno_key1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chart
    ADD CONSTRAINT chart_accno_key1 UNIQUE (accno);


--
-- Name: chat chat_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.chat
    ADD CONSTRAINT chat_pkey PRIMARY KEY (id);


--
-- Name: contact contact_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contact
    ADD CONSTRAINT contact_pkey PRIMARY KEY (id);


--
-- Name: curr curr_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.curr
    ADD CONSTRAINT curr_pkey PRIMARY KEY (curr);


--
-- Name: customer customer_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customer
    ADD CONSTRAINT customer_pkey PRIMARY KEY (id);


--
-- Name: customerlogin customerlogin_login_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customerlogin
    ADD CONSTRAINT customerlogin_login_key UNIQUE (login);


--
-- Name: customerlogin customerlogin_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customerlogin
    ADD CONSTRAINT customerlogin_pkey PRIMARY KEY (id);


--
-- Name: customerlogin customerlogin_session_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customerlogin
    ADD CONSTRAINT customerlogin_session_key UNIQUE (session);


--
-- Name: entities entities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entities
    ADD CONSTRAINT entities_pkey PRIMARY KEY (id);


--
-- Name: financial_year financial_year_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_year
    ADD CONSTRAINT financial_year_pkey PRIMARY KEY (id);


--
-- Name: hierarchy_paths hierarchy_paths_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hierarchy_paths
    ADD CONSTRAINT hierarchy_paths_pkey PRIMARY KEY (ancestor_id, descendant_id);


--
-- Name: lastused lastused_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.lastused
    ADD CONSTRAINT lastused_pkey PRIMARY KEY (id);


--
-- Name: org org_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.org
    ADD CONSTRAINT org_pkey PRIMARY KEY (id);


--
-- Name: paymentmethod paymentmethod_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.paymentmethod
    ADD CONSTRAINT paymentmethod_pkey PRIMARY KEY (id);


--
-- Name: report report_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.report
    ADD CONSTRAINT report_pkey PRIMARY KEY (reportid);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: search_irrelevant_words search_irrelevant_words_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.search_irrelevant_words
    ADD CONSTRAINT search_irrelevant_words_pkey PRIMARY KEY (id);


--
-- Name: tax tax_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tax
    ADD CONSTRAINT tax_pkey PRIMARY KEY (id);


--
-- Name: trf trf_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.trf
    ADD CONSTRAINT trf_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: vat_settlement vat_settlement_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vat_settlement
    ADD CONSTRAINT vat_settlement_pkey PRIMARY KEY (id);


--
-- Name: vendor vendor_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vendor
    ADD CONSTRAINT vendor_pkey PRIMARY KEY (id);


--
-- Name: yearend yearend_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.yearend
    ADD CONSTRAINT yearend_pkey PRIMARY KEY (id);


--
-- Name: acc_trans_chart_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX acc_trans_chart_id_key ON public.acc_trans USING btree (chart_id);


--
-- Name: acc_trans_chart_id_transdate_approved_trans_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX acc_trans_chart_id_transdate_approved_trans_id ON public.acc_trans USING btree (chart_id, transdate, approved, trans_id, amount);


--
-- Name: acc_trans_source_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX acc_trans_source_key ON public.acc_trans USING btree (lower(source));


--
-- Name: acc_trans_trans_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX acc_trans_trans_id_key ON public.acc_trans USING btree (trans_id);


--
-- Name: acc_trans_transdate_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX acc_trans_transdate_key ON public.acc_trans USING btree (transdate);


--
-- Name: ap_quonumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ap_quonumber_key ON public.ap USING btree (lower(quonumber));


--
-- Name: ar_quonumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ar_quonumber_key ON public.ar USING btree (lower(quonumber));


--
-- Name: assembly_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX assembly_id_key ON public.assembly USING btree (id);


--
-- Name: audittrail_trans_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX audittrail_trans_id_key ON public.audittrail USING btree (trans_id);


--
-- Name: cargo_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX cargo_id_key ON public.cargo USING btree (id, trans_id);


--
-- Name: chart_accno_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX chart_accno_key ON public.oldchart USING btree (accno);


--
-- Name: chart_category_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX chart_category_key ON public.oldchart USING btree (category);


--
-- Name: chart_gifi_accno_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX chart_gifi_accno_key ON public.oldchart USING btree (gifi_accno);


--
-- Name: chart_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX chart_id_key ON public.oldchart USING btree (id);


--
-- Name: chart_link_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX chart_link_key ON public.oldchart USING btree (link);


--
-- Name: customer_contact_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX customer_contact_key ON public.customer USING btree (lower((contact)::text));


--
-- Name: customer_customer_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX customer_customer_id_key ON public.customertax USING btree (customer_id);


--
-- Name: customer_customernumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX customer_customernumber_key ON public.customer USING btree (customernumber);


--
-- Name: customer_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX customer_name_key ON public.customer USING btree (lower((name)::text));


--
-- Name: department_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX department_id_key ON public.department USING btree (id);


--
-- Name: employee_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX employee_id_key ON public.employee USING btree (id);


--
-- Name: employee_login_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX employee_login_key ON public.employee USING btree (login);


--
-- Name: employee_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX employee_name_key ON public.employee USING btree (name);


--
-- Name: exchangerate_ct_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX exchangerate_ct_key ON public.exchangerate USING btree (curr, transdate);


--
-- Name: fifo_parts_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX fifo_parts_id ON public.fifo USING btree (parts_id);


--
-- Name: fifo_trans_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX fifo_trans_id ON public.fifo USING btree (trans_id);


--
-- Name: gifi_accno_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX gifi_accno_key ON public.gifi USING btree (accno);


--
-- Name: gl_description_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX gl_description_key ON public.gl USING btree (lower(description));


--
-- Name: gl_employee_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX gl_employee_id_key ON public.gl USING btree (employee_id);


--
-- Name: gl_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX gl_id_key ON public.gl USING btree (id);


--
-- Name: gl_reference_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX gl_reference_key ON public.gl USING btree (reference);


--
-- Name: gl_transdate_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX gl_transdate_key ON public.gl USING btree (transdate);


--
-- Name: idx_entities_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entities_deleted ON public.entities USING btree (deleted_at);


--
-- Name: idx_entities_org; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entities_org ON public.entities USING btree (organization_id);


--
-- Name: idx_entities_org_parent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entities_org_parent ON public.entities USING btree (organization_id, parent_id);


--
-- Name: idx_entities_parent; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entities_parent ON public.entities USING btree (parent_id);


--
-- Name: idx_entities_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_entities_type ON public.entities USING btree (type);


--
-- Name: idx_hierarchy_ancestor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_hierarchy_ancestor ON public.hierarchy_paths USING btree (ancestor_id);


--
-- Name: idx_hierarchy_descendant; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_hierarchy_descendant ON public.hierarchy_paths USING btree (descendant_id);


--
-- Name: idx_hierarchy_paths_depth; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_hierarchy_paths_depth ON public.hierarchy_paths USING btree (depth);


--
-- Name: idx_users_deleted; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted ON public.users USING btree (deleted_at);


--
-- Name: idx_users_entity; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_entity ON public.users USING btree (entity_id);


--
-- Name: idx_users_last_login; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_last_login ON public.users USING btree (last_login_at);


--
-- Name: inventory_invoice_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX inventory_invoice_id ON public.inventory USING btree (invoice_id);


--
-- Name: inventory_parts_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX inventory_parts_id_key ON public.inventory USING btree (parts_id);


--
-- Name: jcitems_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jcitems_id_key ON public.jcitems USING btree (id);


--
-- Name: language_code_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX language_code_key ON public.language USING btree (code);


--
-- Name: makemodel_make_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX makemodel_make_key ON public.makemodel USING btree (lower(make));


--
-- Name: makemodel_model_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX makemodel_model_key ON public.makemodel USING btree (lower(model));


--
-- Name: makemodel_parts_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX makemodel_parts_id_key ON public.makemodel USING btree (parts_id);


--
-- Name: oe_employee_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX oe_employee_id_key ON public.oe USING btree (employee_id);


--
-- Name: oe_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX oe_id_key ON public.oe USING btree (id);


--
-- Name: oe_ordnumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX oe_ordnumber_key ON public.oe USING btree (ordnumber);


--
-- Name: oe_transdate_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX oe_transdate_key ON public.oe USING btree (transdate);


--
-- Name: orderitems_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX orderitems_id_key ON public.orderitems USING btree (id);


--
-- Name: orderitems_trans_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX orderitems_trans_id_key ON public.orderitems USING btree (trans_id);


--
-- Name: parts_description_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX parts_description_key ON public.parts USING btree (lower(description));


--
-- Name: parts_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX parts_id_key ON public.parts USING btree (id);


--
-- Name: parts_partnumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX parts_partnumber_key ON public.parts USING btree (lower(partnumber));


--
-- Name: partscustomer_customer_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX partscustomer_customer_id_key ON public.partscustomer USING btree (customer_id);


--
-- Name: partscustomer_parts_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX partscustomer_parts_id_key ON public.partscustomer USING btree (parts_id);


--
-- Name: partsgroup_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX partsgroup_id_key ON public.partsgroup USING btree (id);


--
-- Name: partsgroup_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX partsgroup_key ON public.partsgroup USING btree (partsgroup);


--
-- Name: partstax_parts_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX partstax_parts_id_key ON public.partstax USING btree (parts_id);


--
-- Name: partsvendor_parts_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX partsvendor_parts_id_key ON public.partsvendor USING btree (parts_id);


--
-- Name: partsvendor_vendor_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX partsvendor_vendor_id_key ON public.partsvendor USING btree (vendor_id);


--
-- Name: pricegroup_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pricegroup_id_key ON public.pricegroup USING btree (id);


--
-- Name: pricegroup_pricegroup_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX pricegroup_pricegroup_key ON public.pricegroup USING btree (pricegroup);


--
-- Name: project_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX project_id_key ON public.project USING btree (id);


--
-- Name: projectnumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX projectnumber_key ON public.project USING btree (projectnumber);


--
-- Name: shipto_trans_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX shipto_trans_id_key ON public.shipto USING btree (trans_id);


--
-- Name: status_trans_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX status_trans_id_key ON public.status USING btree (trans_id);


--
-- Name: translation_trans_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX translation_trans_id_key ON public.translation USING btree (trans_id);


--
-- Name: vendor_contact_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX vendor_contact_key ON public.vendor USING btree (lower((contact)::text));


--
-- Name: vendor_name_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX vendor_name_key ON public.vendor USING btree (lower((name)::text));


--
-- Name: vendor_vendornumber_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX vendor_vendornumber_key ON public.vendor USING btree (vendornumber);


--
-- Name: vendortax_vendor_id_key; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX vendortax_vendor_id_key ON public.vendortax USING btree (vendor_id);


--
-- Name: ap check_department; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON public.ap FOR EACH ROW EXECUTE FUNCTION public.check_department();


--
-- Name: ar check_department; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON public.ar FOR EACH ROW EXECUTE FUNCTION public.check_department();


--
-- Name: gl check_department; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON public.gl FOR EACH ROW EXECUTE FUNCTION public.check_department();


--
-- Name: oe check_department; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER check_department AFTER INSERT OR UPDATE ON public.oe FOR EACH ROW EXECUTE FUNCTION public.check_department();


--
-- Name: customer del_customer; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_customer AFTER DELETE ON public.customer FOR EACH ROW EXECUTE FUNCTION public.del_customer();


--
-- Name: ap del_recurring; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_recurring AFTER DELETE ON public.ap FOR EACH ROW EXECUTE FUNCTION public.del_recurring();


--
-- Name: ar del_recurring; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_recurring AFTER DELETE ON public.ar FOR EACH ROW EXECUTE FUNCTION public.del_recurring();


--
-- Name: gl del_recurring; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_recurring AFTER DELETE ON public.gl FOR EACH ROW EXECUTE FUNCTION public.del_recurring();


--
-- Name: oe del_recurring; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_recurring AFTER DELETE ON public.oe FOR EACH ROW EXECUTE FUNCTION public.del_recurring();


--
-- Name: vendor del_vendor; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_vendor AFTER DELETE ON public.vendor FOR EACH ROW EXECUTE FUNCTION public.del_vendor();


--
-- Name: gl del_yearend; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER del_yearend AFTER DELETE ON public.gl FOR EACH ROW EXECUTE FUNCTION public.del_yearend();


--
-- Name: entities trg_entity_paths; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_entity_paths AFTER INSERT ON public.entities FOR EACH ROW EXECUTE FUNCTION public.update_entity_paths();


--
-- Name: entities trg_update_entity_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_update_entity_timestamp BEFORE UPDATE ON public.entities FOR EACH ROW EXECUTE FUNCTION public.update_timestamps();


--
-- Name: org trg_update_org_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_update_org_timestamp BEFORE UPDATE ON public.org FOR EACH ROW EXECUTE FUNCTION public.update_timestamps();


--
-- Name: users trg_update_user_timestamp; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_update_user_timestamp BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION public.update_timestamps();


--
-- Name: blink_import_process blink_import_process_bank_account_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blink_import_process
    ADD CONSTRAINT blink_import_process_bank_account_id_fkey FOREIGN KEY (bank_account_id) REFERENCES public.bank_account(id);


--
-- Name: blink_import_process_log blink_import_process_log_banking_import_event_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blink_import_process_log
    ADD CONSTRAINT blink_import_process_log_banking_import_event_id_fkey FOREIGN KEY (banking_import_event_id) REFERENCES public.banking_import_event(id);


--
-- Name: blink_import_process_log blink_import_process_log_blink_import_process_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.blink_import_process_log
    ADD CONSTRAINT blink_import_process_log_blink_import_process_id_fkey FOREIGN KEY (blink_import_process_id) REFERENCES public.blink_import_process(id);


--
-- Name: booking_to_settlement booking_to_settlement_settlement_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.booking_to_settlement
    ADD CONSTRAINT booking_to_settlement_settlement_id_fkey FOREIGN KEY (settlement_id) REFERENCES public.vat_settlement(id);


--
-- Name: entities entities_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entities
    ADD CONSTRAINT entities_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.org(id) ON DELETE CASCADE;


--
-- Name: entities entities_parent_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entities
    ADD CONSTRAINT entities_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES public.entities(id) ON DELETE CASCADE;


--
-- Name: financial_year fk_yearend; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.financial_year
    ADD CONSTRAINT fk_yearend FOREIGN KEY (yearend_id) REFERENCES public.yearend(id) ON DELETE SET NULL;


--
-- Name: hierarchy_paths hierarchy_paths_ancestor_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hierarchy_paths
    ADD CONSTRAINT hierarchy_paths_ancestor_id_fkey FOREIGN KEY (ancestor_id) REFERENCES public.entities(id) ON DELETE CASCADE;


--
-- Name: hierarchy_paths hierarchy_paths_descendant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.hierarchy_paths
    ADD CONSTRAINT hierarchy_paths_descendant_id_fkey FOREIGN KEY (descendant_id) REFERENCES public.entities(id) ON DELETE CASCADE;


--
-- Name: users users_entity_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_entity_id_fkey FOREIGN KEY (entity_id) REFERENCES public.entities(id) ON DELETE RESTRICT;


--
-- Name: vr vr_br_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vr
    ADD CONSTRAINT vr_br_id_fkey FOREIGN KEY (br_id) REFERENCES public.br(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

