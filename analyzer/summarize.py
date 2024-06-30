from transformers import pipeline


class Summarizer():

    def __init__(self, model="facebook/bart-large-cnn") -> None:
        self.summarizer = pipeline("summarization", model=model)
    
    def _chunk_text(self, text, max_length=512):
        words = text.split()
        current_chunk = []
        current_length = 0
        
        for word in words:
            if current_length + len(word.split()) <= max_length:
                current_chunk.append(word)
                current_length += len(word.split())
            else:
                yield ' '.join(current_chunk)
                current_chunk = [word]
                current_length = len(word.split())
                
        if current_chunk:
            yield ' '.join(current_chunk)

    def summarize_text(self, text) -> str:
        chunks = list(self._chunk_text(text))
        summaries = self.summarizer(chunks, max_length=150, min_length=30, do_sample=False)
        summarized_text = ' '.join([summary['summary_text'] for summary in summaries])
        return summarized_text