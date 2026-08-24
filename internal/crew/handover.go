package crew

import "example.com/railvolt/internal/model"

func HandoverReceipt(session Session, permit model.Permit) model.Receipt {
	return model.Receipt{PermitID: permit.ID, Worksite: session.Worksite, Owner: session.Owner, Generation: session.Generation, Fencing: session.Fencing}
}
