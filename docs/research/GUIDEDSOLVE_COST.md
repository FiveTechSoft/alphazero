# GuidedSolve bestCost

Prepared, NOT compiled/tested. Replace enqueue-seen with strict bestCost relaxation/reopening. Skip stale heap entries before charging expansion. Hoist logZ once/node. Historical ACKey state identity,softmax objective,victory onpop,API/optimizer unchanged. SolvePath not modified.
Prepared SEARCH_COST_AUDIT=1 TestSearchCostLateCheaperRoute shares production helpers: S->X5,S->Y1,Y->X1,X->G1 requires X2/G3,parentX=Y,one stale discarded. Not executed;no improved resolution/time claim. Softmax branching cost is still a design decision for fixed evaluation,not silently corrected.

Test names/gates refer to the complete preserved validation suite, intentionally not included in this public source branch. No validation result is claimed.
