// Port del bucle caliente: MCTS PUCT sobre tres en raya, sin red
// (prior uniforme), mismo algoritmo y parametros que engine.go:
// 200 partidas x 100 sims, c_puct=1.5, tie-break por primer maximo.
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <time.h>

#define MAXNODES 400000
#define CP 1.5

typedef struct {
    int acts[9], nacts;
    double p[9], w[9];
    int n[9], children[9];
    int expanded;
} Node;

static Node arena[MAXNODES];
static int narena;

static const int WINS[8][3] = {{0,1,2},{3,4,5},{6,7,8},{0,3,6},{1,4,7},{2,5,8},{0,4,8},{2,4,6}};

// terminal: 2 = sigue, 0 = tablas, -1 = perdio el que mueve
static int terminal(const signed char *b) {
    for (int i = 0; i < 8; i++) {
        int s = b[WINS[i][0]] + b[WINS[i][1]] + b[WINS[i][2]];
        if (s == 3 || s == -3) return -1;
    }
    for (int i = 0; i < 9; i++) if (b[i] == 0) return 2;
    return 0;
}

static int new_node(void) { return narena++; }

static void expand(int ni, const signed char *b) {
    Node *nd = &arena[ni];
    int k = 0;
    for (int i = 0; i < 9; i++)
        if (b[i] == 0) { nd->acts[k] = i; nd->p[k] = 0; nd->w[k] = 0; nd->n[k] = 0; nd->children[k] = -1; k++; }
    nd->nacts = k;
    for (int i = 0; i < k; i++) nd->p[i] = 1.0 / k;
    nd->expanded = 1;
}

static int mcts(signed char *b, int player, int sims) {
    int root = new_node();
    memset(&arena[root], 0, sizeof(Node));
    arena[root].expanded = 0;
    expand(root, b);
    for (int s = 0; s < sims; s++) {
        signed char cb[9]; memcpy(cb, b, 9);
        int pl = player;
        int path_nd[16], path_j[16], depth = 0;
        int node = root;
        int t = terminal(cb);
        double v;
        while (t == 2 && arena[node].expanded) {
            int total = 0;
            for (int j = 0; j < arena[node].nacts; j++) total += arena[node].n[j];
            int best = -1; double bs = -1e18;
            for (int j = 0; j < arena[node].nacts; j++) {
                double q = arena[node].n[j] ? arena[node].w[j] / arena[node].n[j] : 0.0;
                double u = CP * arena[node].p[j] * sqrt(total + 1) / (1 + arena[node].n[j]);
                if (q + u > bs) { bs = q + u; best = j; }
            }
            path_nd[depth] = node; path_j[depth] = best; depth++;
            cb[arena[node].acts[best]] = (signed char)pl; pl = -pl;
            if (arena[node].children[best] < 0) {
                arena[node].children[best] = new_node();
                memset(&arena[arena[node].children[best]], 0, sizeof(Node));
            }
            node = arena[node].children[best];
            t = terminal(cb);
        }
        if (t == 2) { v = 0.0; expand(node, cb); }
        else v = (double)t;
        for (int k = depth - 1; k >= 0; k--) {
            v = -v;
            arena[path_nd[k]].n[path_j[k]]++;
            arena[path_nd[k]].w[path_j[k]] += v;
        }
    }
    int best = 0, bn = -1;
    for (int j = 0; j < arena[root].nacts; j++)
        if (arena[root].n[j] > bn) { bn = arena[root].n[j]; best = j; }
    return arena[root].acts[best];
}

int main(void) {
    int games = 200, sims = 100;
    long total = 0;
    struct timespec t0, t1;
    clock_gettime(CLOCK_MONOTONIC, &t0);
    for (int g = 0; g < games; g++) {
        signed char b[9] = {0}; int pl = 1;
        while (terminal(b) == 2) {
            int a = mcts(b, pl, sims);
            total += sims;
            b[a] = (signed char)pl; pl = -pl;
        }
    }
    clock_gettime(CLOCK_MONOTONIC, &t1);
    double el = (t1.tv_sec - t0.tv_sec) + (t1.tv_nsec - t0.tv_nsec) / 1e9;
    printf("[benchttt C] %d partidas x %d sims = %ld sims en %.3fs -> %.0f sims/s (%.1f partidas/s)\n",
           games, sims, total, el, total / el, games / el);
    return 0;
}
