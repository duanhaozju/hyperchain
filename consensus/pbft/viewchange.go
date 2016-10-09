package pbft



func (pbft *pbftProtocal) calcPSet() map[uint64]*ViewChange_PQ {
	pset := make(map[uint64]*ViewChange_PQ)

	for n, p := range pbft.pset {
		pset[n] = p
	}

	for idx, cert := range pbft.certStore {
		if cert.prePrepare == nil {
			continue
		}

		digest := cert.digest
		if !pbft.prepared(digest, idx.v, idx.n) {
			continue
		}

		if p, ok := pset[idx.n]; ok && p.View > idx.v {
			continue
		}

		pset[idx.n] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    digest,
			View:           idx.v,
		}
	}

	return pset
}

func (pbft *pbftProtocal) calcQSet() map[qidx]*ViewChange_PQ {
	qset := make(map[qidx]*ViewChange_PQ)

	for n, q := range pbft.qset {
		qset[n] = q
	}

	for idx, cert := range pbft.certStore {
		if cert.prePrepare == nil {
			continue
		}

		digest := cert.digest
		if !pbft.prePrepared(digest, idx.v, idx.n) {
			continue
		}

		qi := qidx{digest, idx.n}
		if q, ok := qset[qi]; ok && q.View > idx.v {
			continue
		}

		qset[qi] = &ViewChange_PQ{
			SequenceNumber: idx.n,
			BatchDigest:    digest,
			View:           idx.v,
		}
	}

	return qset
}