---
id: qa_ruby
title: Q & A - Ruby specifics
---

Questions and answers based on problems encountered during implementation. 

## Base64 encoded string in Ruby

The inbuilt Base64 library in Ruby is adding some '\n's.

We advice you to use **Base64.strict_encode64()**, which does not add newlines. 

[The Ruby docs](http://ruby-doc.org/stdlib/libdoc/base64/rdoc/classes/Base64.html) are somewhat confusing, the b64encode method is supposed to add a newline for every 60th character, and the example for the encode64 method is actually using the b64encode method. It seems the pack("m") method for the Array class used by encode64 also adds the newlines. I would consider it a design bug that this is not optional.

You could either remove the newlines yourself, or if you're using rails, there's [ActiveSupport::CoreExtensions::Base64::Encoding](http://api.rubyonrails.org/classes/ActiveSupport/CoreExtensions/Base64/Encoding.html) with the encode64s method.

 - [Solution at Stackoverflow](http://stackoverflow.com/questions/2620975/strange-n-in-base64-encoded-string-in-ruby)