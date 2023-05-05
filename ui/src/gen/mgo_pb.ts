// @generated by protoc-gen-es v1.2.0 with parameter "target=ts"
// @generated from file mgo.proto (package mgo.portfolio.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import type { BinaryReadOptions, FieldList, JsonReadOptions, JsonValue, PartialMessage, PlainMessage } from "@bufbuild/protobuf";
import { FieldMask, Message, proto3, Timestamp } from "@bufbuild/protobuf";

/**
 * @generated from message mgo.portfolio.v1.PortfolioCreateMessage
 */
export class PortfolioCreateMessage extends Message<PortfolioCreateMessage> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  constructor(data?: PartialMessage<PortfolioCreateMessage>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.PortfolioCreateMessage";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): PortfolioCreateMessage {
    return new PortfolioCreateMessage().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): PortfolioCreateMessage {
    return new PortfolioCreateMessage().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): PortfolioCreateMessage {
    return new PortfolioCreateMessage().fromJsonString(jsonString, options);
  }

  static equals(a: PortfolioCreateMessage | PlainMessage<PortfolioCreateMessage> | undefined, b: PortfolioCreateMessage | PlainMessage<PortfolioCreateMessage> | undefined): boolean {
    return proto3.util.equals(PortfolioCreateMessage, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.Portfolio
 */
export class Portfolio extends Message<Portfolio> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  constructor(data?: PartialMessage<Portfolio>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.Portfolio";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Portfolio {
    return new Portfolio().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Portfolio {
    return new Portfolio().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Portfolio {
    return new Portfolio().fromJsonString(jsonString, options);
  }

  static equals(a: Portfolio | PlainMessage<Portfolio> | undefined, b: Portfolio | PlainMessage<Portfolio> | undefined): boolean {
    return proto3.util.equals(Portfolio, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.Security
 */
export class Security extends Message<Security> {
  /**
   * Name contains the unique resource name. For a stock or bond, this should be
   * an ISIN.
   *
   * @generated from field: string name = 1;
   */
  name = "";

  /**
   * DisplayName contains the human readable name.
   *
   * @generated from field: string display_name = 2;
   */
  displayName = "";

  /**
   * @generated from field: repeated mgo.portfolio.v1.ListedSecurity listed_on = 4;
   */
  listedOn: ListedSecurity[] = [];

  /**
   * @generated from field: optional string quote_provider = 10;
   */
  quoteProvider?: string;

  constructor(data?: PartialMessage<Security>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.Security";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 2, name: "display_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "listed_on", kind: "message", T: ListedSecurity, repeated: true },
    { no: 10, name: "quote_provider", kind: "scalar", T: 9 /* ScalarType.STRING */, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): Security {
    return new Security().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): Security {
    return new Security().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): Security {
    return new Security().fromJsonString(jsonString, options);
  }

  static equals(a: Security | PlainMessage<Security> | undefined, b: Security | PlainMessage<Security> | undefined): boolean {
    return proto3.util.equals(Security, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.ListedSecurity
 */
export class ListedSecurity extends Message<ListedSecurity> {
  /**
   * @generated from field: string security_name = 1;
   */
  securityName = "";

  /**
   * @generated from field: string ticker = 3;
   */
  ticker = "";

  /**
   * @generated from field: string currency = 4;
   */
  currency = "";

  /**
   * @generated from field: optional float latest_quote = 5;
   */
  latestQuote?: number;

  /**
   * @generated from field: optional google.protobuf.Timestamp latest_quote_timestamp = 6;
   */
  latestQuoteTimestamp?: Timestamp;

  constructor(data?: PartialMessage<ListedSecurity>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.ListedSecurity";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "security_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 3, name: "ticker", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 4, name: "currency", kind: "scalar", T: 9 /* ScalarType.STRING */ },
    { no: 5, name: "latest_quote", kind: "scalar", T: 2 /* ScalarType.FLOAT */, opt: true },
    { no: 6, name: "latest_quote_timestamp", kind: "message", T: Timestamp, opt: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListedSecurity {
    return new ListedSecurity().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListedSecurity {
    return new ListedSecurity().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListedSecurity {
    return new ListedSecurity().fromJsonString(jsonString, options);
  }

  static equals(a: ListedSecurity | PlainMessage<ListedSecurity> | undefined, b: ListedSecurity | PlainMessage<ListedSecurity> | undefined): boolean {
    return proto3.util.equals(ListedSecurity, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.ListSecuritiesRequest
 */
export class ListSecuritiesRequest extends Message<ListSecuritiesRequest> {
  constructor(data?: PartialMessage<ListSecuritiesRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.ListSecuritiesRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListSecuritiesRequest {
    return new ListSecuritiesRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListSecuritiesRequest {
    return new ListSecuritiesRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListSecuritiesRequest {
    return new ListSecuritiesRequest().fromJsonString(jsonString, options);
  }

  static equals(a: ListSecuritiesRequest | PlainMessage<ListSecuritiesRequest> | undefined, b: ListSecuritiesRequest | PlainMessage<ListSecuritiesRequest> | undefined): boolean {
    return proto3.util.equals(ListSecuritiesRequest, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.ListSecuritiesResponse
 */
export class ListSecuritiesResponse extends Message<ListSecuritiesResponse> {
  /**
   * @generated from field: repeated mgo.portfolio.v1.Security securities = 1;
   */
  securities: Security[] = [];

  constructor(data?: PartialMessage<ListSecuritiesResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.ListSecuritiesResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "securities", kind: "message", T: Security, repeated: true },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): ListSecuritiesResponse {
    return new ListSecuritiesResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): ListSecuritiesResponse {
    return new ListSecuritiesResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): ListSecuritiesResponse {
    return new ListSecuritiesResponse().fromJsonString(jsonString, options);
  }

  static equals(a: ListSecuritiesResponse | PlainMessage<ListSecuritiesResponse> | undefined, b: ListSecuritiesResponse | PlainMessage<ListSecuritiesResponse> | undefined): boolean {
    return proto3.util.equals(ListSecuritiesResponse, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.GetSecurityRequest
 */
export class GetSecurityRequest extends Message<GetSecurityRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  constructor(data?: PartialMessage<GetSecurityRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.GetSecurityRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): GetSecurityRequest {
    return new GetSecurityRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): GetSecurityRequest {
    return new GetSecurityRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): GetSecurityRequest {
    return new GetSecurityRequest().fromJsonString(jsonString, options);
  }

  static equals(a: GetSecurityRequest | PlainMessage<GetSecurityRequest> | undefined, b: GetSecurityRequest | PlainMessage<GetSecurityRequest> | undefined): boolean {
    return proto3.util.equals(GetSecurityRequest, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.CreateSecurityRequest
 */
export class CreateSecurityRequest extends Message<CreateSecurityRequest> {
  /**
   * @generated from field: mgo.portfolio.v1.Security security = 1;
   */
  security?: Security;

  constructor(data?: PartialMessage<CreateSecurityRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.CreateSecurityRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "security", kind: "message", T: Security },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): CreateSecurityRequest {
    return new CreateSecurityRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): CreateSecurityRequest {
    return new CreateSecurityRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): CreateSecurityRequest {
    return new CreateSecurityRequest().fromJsonString(jsonString, options);
  }

  static equals(a: CreateSecurityRequest | PlainMessage<CreateSecurityRequest> | undefined, b: CreateSecurityRequest | PlainMessage<CreateSecurityRequest> | undefined): boolean {
    return proto3.util.equals(CreateSecurityRequest, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.UpdateSecurityRequest
 */
export class UpdateSecurityRequest extends Message<UpdateSecurityRequest> {
  /**
   * @generated from field: mgo.portfolio.v1.Security security = 1;
   */
  security?: Security;

  /**
   * @generated from field: google.protobuf.FieldMask update_mask = 2;
   */
  updateMask?: FieldMask;

  constructor(data?: PartialMessage<UpdateSecurityRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.UpdateSecurityRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "security", kind: "message", T: Security },
    { no: 2, name: "update_mask", kind: "message", T: FieldMask },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): UpdateSecurityRequest {
    return new UpdateSecurityRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): UpdateSecurityRequest {
    return new UpdateSecurityRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): UpdateSecurityRequest {
    return new UpdateSecurityRequest().fromJsonString(jsonString, options);
  }

  static equals(a: UpdateSecurityRequest | PlainMessage<UpdateSecurityRequest> | undefined, b: UpdateSecurityRequest | PlainMessage<UpdateSecurityRequest> | undefined): boolean {
    return proto3.util.equals(UpdateSecurityRequest, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.DeleteSecurityRequest
 */
export class DeleteSecurityRequest extends Message<DeleteSecurityRequest> {
  /**
   * @generated from field: string name = 1;
   */
  name = "";

  constructor(data?: PartialMessage<DeleteSecurityRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.DeleteSecurityRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): DeleteSecurityRequest {
    return new DeleteSecurityRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): DeleteSecurityRequest {
    return new DeleteSecurityRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): DeleteSecurityRequest {
    return new DeleteSecurityRequest().fromJsonString(jsonString, options);
  }

  static equals(a: DeleteSecurityRequest | PlainMessage<DeleteSecurityRequest> | undefined, b: DeleteSecurityRequest | PlainMessage<DeleteSecurityRequest> | undefined): boolean {
    return proto3.util.equals(DeleteSecurityRequest, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.TriggerQuoteUpdateRequest
 */
export class TriggerQuoteUpdateRequest extends Message<TriggerQuoteUpdateRequest> {
  /**
   * @generated from field: string security_name = 1;
   */
  securityName = "";

  constructor(data?: PartialMessage<TriggerQuoteUpdateRequest>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.TriggerQuoteUpdateRequest";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
    { no: 1, name: "security_name", kind: "scalar", T: 9 /* ScalarType.STRING */ },
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TriggerQuoteUpdateRequest {
    return new TriggerQuoteUpdateRequest().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TriggerQuoteUpdateRequest {
    return new TriggerQuoteUpdateRequest().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TriggerQuoteUpdateRequest {
    return new TriggerQuoteUpdateRequest().fromJsonString(jsonString, options);
  }

  static equals(a: TriggerQuoteUpdateRequest | PlainMessage<TriggerQuoteUpdateRequest> | undefined, b: TriggerQuoteUpdateRequest | PlainMessage<TriggerQuoteUpdateRequest> | undefined): boolean {
    return proto3.util.equals(TriggerQuoteUpdateRequest, a, b);
  }
}

/**
 * @generated from message mgo.portfolio.v1.TriggerQuoteUpdateResponse
 */
export class TriggerQuoteUpdateResponse extends Message<TriggerQuoteUpdateResponse> {
  constructor(data?: PartialMessage<TriggerQuoteUpdateResponse>) {
    super();
    proto3.util.initPartial(data, this);
  }

  static readonly runtime: typeof proto3 = proto3;
  static readonly typeName = "mgo.portfolio.v1.TriggerQuoteUpdateResponse";
  static readonly fields: FieldList = proto3.util.newFieldList(() => [
  ]);

  static fromBinary(bytes: Uint8Array, options?: Partial<BinaryReadOptions>): TriggerQuoteUpdateResponse {
    return new TriggerQuoteUpdateResponse().fromBinary(bytes, options);
  }

  static fromJson(jsonValue: JsonValue, options?: Partial<JsonReadOptions>): TriggerQuoteUpdateResponse {
    return new TriggerQuoteUpdateResponse().fromJson(jsonValue, options);
  }

  static fromJsonString(jsonString: string, options?: Partial<JsonReadOptions>): TriggerQuoteUpdateResponse {
    return new TriggerQuoteUpdateResponse().fromJsonString(jsonString, options);
  }

  static equals(a: TriggerQuoteUpdateResponse | PlainMessage<TriggerQuoteUpdateResponse> | undefined, b: TriggerQuoteUpdateResponse | PlainMessage<TriggerQuoteUpdateResponse> | undefined): boolean {
    return proto3.util.equals(TriggerQuoteUpdateResponse, a, b);
  }
}

