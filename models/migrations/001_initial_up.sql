-- 001_create_users_table.sql

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Define user roles enum
CREATE TYPE user_role AS ENUM ('applicant');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    reg_num VARCHAR(15) NOT NULL,
    hashed_password TEXT NOT NULL,
    reset_token TEXT,
    reset_token_expires_at TIMESTAMP WITH TIME ZONE,
    role user_role NOT NULL DEFAULT 'applicant',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Define department enum for questions
CREATE TYPE department AS ENUM ('technical', 'design', 'management', 'social');

-- Create questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    department department NOT NULL,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Create applications table
CREATE TABLE IF NOT EXISTS applications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    department department NOT NULL,
    submitted BOOLEAN NOT NULL DEFAULT FALSE,
    chickened_out BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, department)
);

-- Create trigger for applications updated_at
CREATE TRIGGER update_applications_updated_at
BEFORE UPDATE ON applications
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create answers table
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(application_id, question_id)
);

-- Create trigger for answers updated_at
CREATE TRIGGER update_answers_updated_at
BEFORE UPDATE ON answers
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Insert questions for Social department
INSERT INTO questions (department, title, body) VALUES
('social', 'What tools or apps do you use (or want to learn) for editing, designing?', 'Share your experience or interest in creative tools, focusing on editing and design. Mention any software or apps familiar to you or those you want to explore. The answer can describe general preferences or goals for learning new tools. Focus is on your adaptability and openness to technology.'),
('social', 'Can you share the link of a post, reel you''ve created (for a class, event, personal page, or another club)?', 'Provide an example or examples of your social media work. It can be from any relevant context like school, events, or personal use. The answer should highlight your involvement in content creation. Details about the post content or purpose are optional.'),
('social', 'If you were asked to create a 30-second reel for Instagram using event clips, how would you make it engaging and trendy?', 'Describe your approach to making short social media videos appealing. Focus on ideas for engaging viewers quickly. Mention any general techniques or trends you might consider. The answer could include how to balance creativity and audience interest.'),
('social', 'If you had to run the club''s Instagram using only stories for one week, what kind of story content would you create each day to keep students hooked?', 'Outline a general plan for daily content using Instagram stories. The focus is on maintaining variety and interest over a week. Mention possible content types or engagement ideas. The goal is to show how you might keep the audience involved.'),
('social', 'How do you balance being creative while still representing the club professionally?', 'Explain your thoughts on combining creativity with a professional image. Mention how you might consider the club''s values or audience. The answer can focus on finding a suitable tone and style. Describe balancing innovation and appropriateness.'),
('social', 'If you were given full creative control of our social media for the club''s flagship event, what theme would you choose, what type of content would you create, and how would you plan its execution from pre-event hype to post-event highlights?', 'Talk about your general approach to leading social media for a big event. Include thoughts on theme selection and content types. Explain how you might plan and organize different phases. Focus is on a broad overview rather than specific details.'),
('social', 'If our club has multiple teams (tech, design, management), how would you ensure the social media strategy represents everyone fairly?', 'Discuss how you might approach balancing content for different teams. Mention ideas for inclusive representation without naming specific tactics. The goal is to show awareness of fairness and collaboration. Emphasize keeping all groups engaged.'),
('social', 'Our event clashes with another club''s event. How would you make our promotions stand out and still grab attention?', 'Describe general ways to differentiate your event promotions from competitors. Focus on highlighting strengths or unique features. Mention possible creative or timing strategies. The answer should emphasize respectful competition.'),
('social', 'How would you handle negative or inappropriate comments on our posts?', 'Explain your overall approach to managing difficult comments. Mention key principles like respect and professionalism. The answer can cover basic steps like monitoring and responding or escalating. Focus on preserving the club''s reputation.'),
('social', 'One of our posts unintentionally goes viral but is being used for memes unrelated to our club. Would you ride the trend or control the narrative? Explain.', 'Share your thoughts on responding to unexpected viral content. Mention the pros and cons of engaging with or steering the narrative. The answer could show how you weigh opportunities against risks. Focus on thoughtful decision-making.');

-- Insert questions for Management department
INSERT INTO questions (department, title, body) VALUES
('management', 'What strategies would you use to motivate a team that is struggling to meet deadlines?', 'Describe your general approach to improving team motivation and productivity. Focus on fostering a positive environment and understanding individual challenges. Mention the importance of support and communication. Keep the answer broad and adaptable.'),
('management', 'Describe your process for organizing an event or project from start to finish.', 'Outline your typical method for planning and execution. Highlight key stages without specifics, such as preparation, coordination, and follow-up. Emphasize ensuring timely completion and team collaboration. Keep open to various workflows or tools.'),
('management', 'How would you handle a situation where a team member is not meeting expectations?', 'Discuss your approach to addressing performance concerns tactfully. Focus on communication and offering help rather than confrontation. Mention the importance of maintaining team harmony and encouraging improvement. Keep it flexible to different scenarios.'),
('management', 'Imagine you have been made the student coordinator for an event and need help from departments like tech, social media, and HR. What would be your approach for delegating tasks and ensuring completion?', 'Explain your general delegation strategy and how you would coordinate multiple teams. Mention the importance of clear communication and accountability. Focus on balancing workloads and following up on progress. Keep explanation open to different leadership styles.'),
('management', 'If a project or event doesn''t go as planned and there is a hurdle at the eleventh hour, how would you handle it and what lessons would you take away?', 'Talk about staying calm and adaptable during unexpected challenges. Emphasize problem-solving and teamwork. Mention learning from setbacks to improve future planning. Keep the description broad and centered on resilience.'),
('management', 'Suppose you''ve been made a lead for a portion of the event, say logistics, but the members in your department are failing to show up and meet deadlines while you''re getting a lot of pressure from the board to get this done. What path would you take? What if this behaviour is repeated by your team in the next event?', 'Discuss how you would handle accountability and responsibility under pressure. Emphasize balancing diplomacy with firmness - addressing the team''s lack of participation while ensuring the event is not compromised. Highlight the importance of communication, redistributing work if necessary, and escalating when repeated behavior affects the club''s goals. Keep the response flexible to show leadership, resilience, and fairness.'),
('management', 'If the club''s event is experiencing low registrations, how would you increase participation?', 'Explain your general strategies for boosting interest and engagement. Mention outreach, promotions, and possibly incentives. Focus on responsiveness and adaptability to changing situations. Keep the description broad and creative.'),
('management', 'How would you respectfully disagree with suggestions made by a senior club member, especially as a new recruit?', 'Talk about the importance of diplomacy, active listening, and timing. Mention respectfully sharing alternative ideas and seeking common ground. Emphasize maintaining good relationships and open communication. Keep the response general and professional.'),
('management', 'In your view, what does it truly mean to be part of a community? How do you personally experience belonging, and what principles do you believe are most important in building and managing a strong community?', 'We encourage you to share what being part of a community means to you. Think about values like inclusivity, openness, and collaboration. Tell us how you feel a sense of belonging and what principles you believe matter most in building and managing a supportive community where everyone feels welcome and valued.'),
('management', 'To end on a reflective note: what are three reasons you believe we should not select you for this club?', 'We encourage self-awareness and honesty. Our aim is to see how you reflect on your weaknesses or limitations, while also showing you are open to growth. Keep it flexible to show maturity, accountability, and willingness to improve.');

-- Insert questions for Design department
INSERT INTO questions (department, title, body) VALUES
('design', 'Your go-to design tool & why?', 'Share your preferred creative software or tool broadly. Explain what draws you to it or how it supports your work. The answer can be based on features, ease of use, or personal comfort. Keep the explanation open and reflective.'),
('design', 'One UI/UX project you''ve done — what was your role?', 'Describe an example of a design project you participated in. Focus on your general contribution without going into technical specifics. Mention teamwork or individual role with flexibility. Emphasize involvement over detailed outcomes.'),
('design', 'Ever redesigned an app/website just for fun? Which one?', 'Talk about any personal or informal design exercises you have done. Share what inspired you without focusing on the final product. Emphasize creativity and practice. Keep the explanation general.'),
('design', 'If you made our next event page, what 3 things would you add?', 'Suggest broad ideas or features you think would improve a webpage. Focus on potential user benefits or engagement. Avoid technical details or exact implementations. Keep ideas open and adaptable.'),
('design', 'One thing you''d improve on our current website?', 'Highlight an area you think could use enhancement or change. Keep the comment high-level rather than highly specific. Show awareness of general user experience or aesthetic considerations. Encourage open interpretation.'),
('design', 'Honest take: how''s the design of our Insta posts?', 'Give general feedback without going into specifics about individual posts. Discuss overall impressions or feelings. Mention balance of aesthetics and messaging. Keep the perspective open and constructive.'),
('design', 'Why IEEE Compsoc?', 'Explain your broad reasons for interest in the community or club. Focus on values, opportunities, or personal alignment. Keep the answer authentic but adaptable. Avoid overly detailed or specific points.'),
('design', 'Which app/website do you love just for its design?', 'Name a design you find inspiring or enjoyable. Share reasons in a broad sense, like usability or visual appeal. Avoid detailed critique or technical jargon. Keep it a personal favorite with flexible interpretation.'),
('design', 'Got tough feedback on a design? How did you handle it?', 'Describe your general approach to receiving and responding to critique. Emphasize openness, learning, and professional growth. Focus on attitude rather than exact actions. Keep the explanation broad and positive.'),
('design', 'What unique vibe would you bring to our designs?', 'Talk about the kind of creative energy or style you aim to contribute. Focus on general attributes like innovation, freshness, or collaboration. Avoid technical or highly specific design philosophies. Keep it open and aspirational.');

-- Insert questions for Technical department
INSERT INTO questions (department, title, body) VALUES
('technical', 'What tech stack do you primarily work with and how long have you been using it?', 'Share your main programming languages, frameworks, and tools that you use regularly. Mention your experience level with each and how long you''ve been working with them. Include any technologies you''re currently learning or want to explore. Focus on demonstrating your technical foundation and growth mindset.'),
('technical', 'Describe a challenging technical problem you''ve solved. What was your approach?', 'Walk us through a difficult coding or technical challenge you encountered. Explain your problem-solving process, including how you researched, debugged, or approached the solution. Highlight your analytical thinking and persistence. The focus is on your methodology rather than the specific technical details.'),
('technical', 'How do you approach learning a new technology or programming language?', 'Describe your general strategy for picking up new technical skills. Mention your preferred learning resources, how you practice, and how you stay updated with technology trends. Show your adaptability and commitment to continuous learning in the rapidly evolving tech field.'),
('technical', 'Tell us about a time you worked on a team project. What was your role and how did you collaborate?', 'Share an experience of working with others on a technical project. Describe your specific contributions, how you communicated with team members, and how you handled any challenges that arose. Emphasize teamwork, communication skills, and your ability to work in collaborative environments.'),
('technical', 'What''s your experience with version control systems like Git? How do you use them in your projects?', 'Explain your familiarity with version control and collaborative development practices. Mention how you organize your code, handle branching, and work with others on shared codebases. Show your understanding of professional development workflows and team coding practices.'),
('technical', 'Scenario: You''re working on a club website that suddenly crashes during a major event registration. The deadline is in 2 hours and hundreds of students are trying to register. How would you handle this crisis?', 'Describe your approach to handling a high-pressure technical emergency. Focus on prioritization, quick diagnosis, communication with stakeholders, and emergency response. Show your ability to stay calm under pressure, think systematically, and balance quick fixes with proper solutions. Demonstrate crisis management and technical leadership skills.'),
('technical', 'How do you ensure your code is clean, maintainable, and follows best practices?', 'Explain your approach to writing quality code. Mention coding standards, documentation, testing, and code organization practices you follow. Show your understanding of software craftsmanship and long-term maintainability. Demonstrate your commitment to professional development practices.'),
('technical', 'What''s your experience with databases? Which ones have you worked with and in what context?', 'Share your experience with database systems, whether SQL or NoSQL. Mention specific databases you''ve used, types of applications you''ve built, and your understanding of data modeling. Include any experience with database design, optimization, or administration. Show your backend and data management capabilities.'),
('technical', 'Scenario: A team member pushes code that breaks the main branch right before a demo to potential sponsors. The team is blaming each other and morale is low. As someone who noticed the issue, how would you help resolve both the technical problem and team conflict?', 'Address both technical crisis management and team dynamics. Show your problem-solving approach for the code issue, conflict resolution skills, and ability to maintain team cohesion under stress. Demonstrate leadership qualities, empathy, and professional maturity in handling both technical and interpersonal challenges in a team setting.'),
('technical', 'What motivates you to pursue technical projects, and how do you see technology impacting our club''s goals?', 'Share your passion for technology and how it aligns with contributing to a computer science community. Explain what drives your interest in coding and technical problem-solving. Discuss how you envision using your technical skills to enhance club activities, events, or member experiences. Show your vision and enthusiasm for applying technology meaningfully.');

-- Optional: Create a super admin user (update the email/password as needed)
-- INSERT INTO users (id, full_name, email, verified, reg_num, hashed_password, role)
-- VALUES (
--     uuid_generate_v4(),
--     'Root',
--     'admin@comp.socks',
--     true,
--     '+91 9898888110',
--     '$2a$10$Q8Ltxi7JDz.VJydOo1d73eorls8XOL1OihDfSMwiZo.mJ0fNip.1C',
--     'super_admin'
-- )
