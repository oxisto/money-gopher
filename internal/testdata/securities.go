package testdata

import (
	"time"

	"github.com/google/uuid"
	moneygopher "github.com/oxisto/money-gopher"
	"github.com/oxisto/money-gopher/currency"
	"github.com/oxisto/money-gopher/internal/testing/quotetest"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/portfolio/accounts"
	"github.com/oxisto/money-gopher/portfolio/events"
)

// TestSecurity is a test security.
var TestSecurity = &persistence.Security{
	ID:            "DE1234567890",
	DisplayName:   "My Security",
	QuoteProvider: moneygopher.Ref(quotetest.QuoteProviderStatic),
}

// TestListedSecurity is a listed security for [TestSecurity] that has a ticker
// "TICK" and currency "USD".
var TestListedSecurity = &persistence.ListedSecurity{
	SecurityID: TestSecurity.ID,
	Ticker:     "TICK",
	Currency:   "USD",
}

// TestCreateSecurityParams is a test security creation parameter.
var TestCreateSecurityParams = persistence.CreateSecurityParams{
	ID:            TestSecurity.ID,
	DisplayName:   TestSecurity.DisplayName,
	QuoteProvider: TestSecurity.QuoteProvider,
}

// TestUpsertListedSecurityParams is a test listed security upsert parameter.
var TestUpsertListedSecurityParams = persistence.UpsertListedSecurityParams{
	SecurityID: TestSecurity.ID,
	Ticker:     TestListedSecurity.Ticker,
	Currency:   TestListedSecurity.Currency,
}

// TestBankAccount is a test bank account.
var TestBankAccount = &persistence.Account{
	ID:          "myaccount",
	DisplayName: "My Account",
	Type:        accounts.AccountTypeBank,
}

// TestBrokerageAccount is a test security account.
var TestBrokerageAccount = &persistence.Account{
	ID:          "mybrokerage",
	DisplayName: "My Brokerage",
	Type:        accounts.AccountTypeBrokerage,
}

// TestCreateBankAccountParams is a test bank account creation parameter for
// [TestBankAccount].
var TestCreateBankAccountParams = persistence.CreateAccountParams{
	ID:          TestBankAccount.ID,
	DisplayName: TestBankAccount.DisplayName,
	Type:        TestBankAccount.Type,
}

// TestCreateBrokerageAccountParams is a test brokerage account creation
// parameter for [TestBrokerageAccount].
var TestCreateBrokerageAccountParams = persistence.CreateAccountParams{
	ID:          TestBrokerageAccount.ID,
	DisplayName: TestBrokerageAccount.DisplayName,
	Type:        TestBrokerageAccount.Type,
}

// TestBuyTransaction is a test buy transaction of [TestSecurity]. The buy is
// initiated from [TestBankAccount] and the stocks are deposited in
// [TestBrokerageAccount].
var TestBuyTransaction = &persistence.Transaction{
	ID:                   uuid.NewString(),
	SourceAccountID:      &TestBankAccount.ID,
	DestinationAccountID: &TestBrokerageAccount.ID,
	Time:                 time.Now(),
	Type:                 events.PortfolioEventTypeBuy,
	Amount:               100,
	SecurityID:           &TestSecurity.ID,
	Price:                currency.Value(100),
}

// TestCreateBuyTransactionParams is a test buy transaction creation parameter
// for [TestBuyTransaction].
var TestCreateBuyTransactionParams = persistence.CreateTransactionParams{
	ID:                   TestBuyTransaction.ID,
	SourceAccountID:      TestBuyTransaction.SourceAccountID,
	DestinationAccountID: TestBuyTransaction.DestinationAccountID,
	Time:                 TestBuyTransaction.Time,
	Type:                 TestBuyTransaction.Type,
	Amount:               TestBuyTransaction.Amount,
	SecurityID:           TestBuyTransaction.SecurityID,
	Price:                TestBuyTransaction.Price,
}

// TestDepositTransaction is a test deposit transaction. The deposit is made to
// [TestBankAccount].
var TestDepositTransaction = &persistence.Transaction{
	ID:                   uuid.NewString(),
	DestinationAccountID: &TestBankAccount.ID,
	Time:                 time.Now(),
	Type:                 events.PortfolioEventTypeBuy,
	Amount:               1,
	Price:                currency.Value(100),
}

// TestCreateDepositTransactionParams is a test deposit transaction creation
// parameter for [TestDepositTransaction].
var TestCreateDepositTransactionParams = persistence.CreateTransactionParams{
	ID:                   TestDepositTransaction.ID,
	DestinationAccountID: TestDepositTransaction.DestinationAccountID,
	Time:                 TestDepositTransaction.Time,
	Type:                 TestDepositTransaction.Type,
	Amount:               TestDepositTransaction.Amount,
	Price:                TestDepositTransaction.Price,
}

// TestPortfolio is a test portfolio.
var TestPortfolio = &persistence.Portfolio{
	ID:          "myportfolio",
	DisplayName: "My Portfolio",
}

// TestCreatePortfolioParams is a test portfolio creation parameter for
// [TestPortfolio].
var TestCreatePortfolioParams = persistence.CreatePortfolioParams{
	ID:          TestPortfolio.ID,
	DisplayName: TestPortfolio.DisplayName,
}

// TestAddAccountToPortfolioParams is a test account addition parameter for
// [TestBankAccount] to [TestPortfolio].
var TestAddAccountToPortfolioParams = persistence.AddAccountToPortfolioParams{
	PortfolioID: TestPortfolio.ID,
	AccountID:   TestBankAccount.ID,
}