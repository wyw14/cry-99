package crew

import "example.com/railvolt/internal/model"

func ReceiptAccepted(receipt model.Receipt, permit model.Permit) bool {
	return receipt.PermitID == permit.ID && receipt.Owner == permit.Owner && receipt.Generation == permit.Generation && receipt.Fencing != ""
}
