macro_rules! ord {
    ($c:expr) => { $c as usize - 'a' as usize }
}

trait Traverse<'input, 'trie> {
    fn add(&'trie mut self, s: &'input [u8]);
    fn score_acc(&'trie self, s: &'input [u8], acc: i32) -> i32;
    fn score(&'trie self, s: &'input [u8]) -> i32 {
        return self.score_acc(s, 0)
    }
}

#[derive(Default)]
enum Node<'input> {
    #[default]
    Empty,
    Pending(&'input [u8]),
    Normal { edges: [Option<Box<Edge<'input>>>; ord!('z') + 1] },
}

impl<'input, 'trie> Traverse<'input, 'trie> for Node<'input> {
    fn add(&'trie mut self, s: &'input [u8]) {
        match self {
            Node::Empty => {
                *self = Node::Pending(s);
            }
            Node::Pending(t) => {
                let mut normal = Node::Normal { edges: Default::default() };
                normal.add(t);
                normal.add(s);
                *self = normal;
            }
            Node::Normal { edges } => {
                let Some((&c, tail)) = s.split_first() else { return };
                let edge = edges[ord!(c)].get_or_insert_with(|| Box::default());
                edge.add(tail);
            }
        }
    }
    fn score_acc(&'trie self, s: &'input [u8], acc: i32) -> i32 {
        match self {
            Node::Empty => { return acc; }
            Node::Pending(t) => { return acc + t.len() as i32; }
            Node::Normal { edges } => {
                let Some((&c, tail)) = s.split_first() else { return acc };
                let Some(ref edge) = edges[ord!(c)] else { return acc };
                return edge.score_acc(tail, acc)
            }
        }
    }
}

#[derive(Default)]
struct Edge<'input> {
    node: Node<'input>,
    count: i32,
}

impl<'input, 'trie> Traverse<'input, 'trie> for Edge<'input> {
    fn add(&'trie mut self, s: &'input [u8]) {
        self.count += 1;
        self.node.add(s);
    }
    fn score_acc(&'trie self, s: &'input [u8], acc: i32) -> i32 {
        self.node.score_acc(s, acc + self.count)
    }
}

impl Solution {
    pub fn sum_prefix_scores(words: Vec<String>) -> Vec<i32> {
        let mut root = Node::default();
        for w in &words {
            root.add(w.as_bytes());
        }
        return words.iter().map(|w| root.score(w.as_bytes())).collect()
    }
}
