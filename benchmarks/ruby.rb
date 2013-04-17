require 'chunky_png'

base_path = File.expand_path('../../fixtures/large/base.png', __FILE__)
target_path = File.expand_path('../../fixtures/large/target.png', __FILE__)
base = ChunkyPNG::Image.from_file(base_path)
target = ChunkyPNG::Image.from_file(target_path)

puts base.area
puts target.area
