// @generated by protoc-gen-connect-es v0.8.6 with parameter "target=ts"
// @generated from file mgo.proto (package mgo.portfolio.v1, syntax proto3)
/* eslint-disable */
// @ts-nocheck

import { CreatePortfolioRequest, CreatePortfolioTransactionRequest, CreateSecurityRequest, DeletePortfolioRequest, DeletePortfolioTransactionRequest, DeleteSecurityRequest, GetPortfolioSnapshotRequest, GetSecurityRequest, ListPortfolioRequest, ListPortfoliosResponse, ListPortfolioTransactionsRequest, ListPortfolioTransactionsResponse, ListSecuritiesRequest, ListSecuritiesResponse, Portfolio, PortfolioEvent, PortfolioSnapshot, Security, TriggerQuoteUpdateRequest, TriggerQuoteUpdateResponse, UpdatePortfolioRequest, UpdatePortfolioTransactionRequest, UpdateSecurityRequest } from "./mgo_pb.js";
import { Empty, MethodIdempotency, MethodKind } from "@bufbuild/protobuf";

/**
 * @generated from service mgo.portfolio.v1.PortfolioService
 */
export const PortfolioService = {
  typeName: "mgo.portfolio.v1.PortfolioService",
  methods: {
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.CreatePortfolio
     */
    createPortfolio: {
      name: "CreatePortfolio",
      I: CreatePortfolioRequest,
      O: Portfolio,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.ListPortfolio
     */
    listPortfolio: {
      name: "ListPortfolio",
      I: ListPortfolioRequest,
      O: ListPortfoliosResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.UpdatePortfolio
     */
    updatePortfolio: {
      name: "UpdatePortfolio",
      I: UpdatePortfolioRequest,
      O: Portfolio,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.DeletePortfolio
     */
    deletePortfolio: {
      name: "DeletePortfolio",
      I: DeletePortfolioRequest,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.GetPortfolioSnapshot
     */
    getPortfolioSnapshot: {
      name: "GetPortfolioSnapshot",
      I: GetPortfolioSnapshotRequest,
      O: PortfolioSnapshot,
      kind: MethodKind.Unary,
    idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.CreatePortfolioTransaction
     */
    createPortfolioTransaction: {
      name: "CreatePortfolioTransaction",
      I: CreatePortfolioTransactionRequest,
      O: PortfolioEvent,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.ListPortfolioTransactions
     */
    listPortfolioTransactions: {
      name: "ListPortfolioTransactions",
      I: ListPortfolioTransactionsRequest,
      O: ListPortfolioTransactionsResponse,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.UpdatePortfolioTransaction
     */
    updatePortfolioTransaction: {
      name: "UpdatePortfolioTransaction",
      I: UpdatePortfolioTransactionRequest,
      O: PortfolioEvent,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.PortfolioService.DeletePortfolioTransaction
     */
    deletePortfolioTransaction: {
      name: "DeletePortfolioTransaction",
      I: DeletePortfolioTransactionRequest,
      O: Empty,
      kind: MethodKind.Unary,
    },
  }
} as const;

/**
 * @generated from service mgo.portfolio.v1.SecuritiesService
 */
export const SecuritiesService = {
  typeName: "mgo.portfolio.v1.SecuritiesService",
  methods: {
    /**
     * @generated from rpc mgo.portfolio.v1.SecuritiesService.ListSecurities
     */
    listSecurities: {
      name: "ListSecurities",
      I: ListSecuritiesRequest,
      O: ListSecuritiesResponse,
      kind: MethodKind.Unary,
    idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.SecuritiesService.GetSecurity
     */
    getSecurity: {
      name: "GetSecurity",
      I: GetSecurityRequest,
      O: Security,
      kind: MethodKind.Unary,
    idempotency: MethodIdempotency.NoSideEffects,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.SecuritiesService.CreateSecurity
     */
    createSecurity: {
      name: "CreateSecurity",
      I: CreateSecurityRequest,
      O: Security,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.SecuritiesService.UpdateSecurity
     */
    updateSecurity: {
      name: "UpdateSecurity",
      I: UpdateSecurityRequest,
      O: Security,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.SecuritiesService.DeleteSecurity
     */
    deleteSecurity: {
      name: "DeleteSecurity",
      I: DeleteSecurityRequest,
      O: Empty,
      kind: MethodKind.Unary,
    },
    /**
     * @generated from rpc mgo.portfolio.v1.SecuritiesService.TriggerSecurityQuoteUpdate
     */
    triggerSecurityQuoteUpdate: {
      name: "TriggerSecurityQuoteUpdate",
      I: TriggerQuoteUpdateRequest,
      O: TriggerQuoteUpdateResponse,
      kind: MethodKind.Unary,
    },
  }
} as const;

